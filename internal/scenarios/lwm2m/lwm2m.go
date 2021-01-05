package lwm2m

import (
	"net/http"
	"oes-acceptance-testing/v1/internal/tester"
	"oes-acceptance-testing/v1/pkg/helpers/conf"
	"oes-acceptance-testing/v1/pkg/helpers/logger"
	"oes-acceptance-testing/v1/pkg/helpers/random"
	"oes-acceptance-testing/v1/pkg/nbclient"
	"sync"
	"time"
)

const (
	suiteName = "LWM2M测试套件"
	td        = true
)

var wg sync.WaitGroup

func RunLwm2mTest(resultChan chan tester.TestResult, onHold bool) {
	var err error
	// initialize logger with suite name
	le, f := logger.NewLogger(suiteName)
	defer f.Close()

	// initialize tester
	t := &tester.Tester{
		SuiteName: suiteName,
		Logger:    le,
		Teardown:  td,
		Client:    &http.Client{},
	}
	// set up teardown
	if t.Teardown == true {
		defer tearDown(t, resultChan)
	}

	// recover if panic, then proceed to teardown
	defer t.Recover()

	// 创建LWM2M产品
	cp := tester.CreateDmpProductPl{
		Name:                 "lwm2m_" + random.RandString(),
		Description:          "some description",
		ProtocolType:         tester.PROTOCOL_LWM2M,
		NodeType:             tester.NODETYPE_DEVICE,
		NetworkMethod:        1,
		AuthenticationMethod: tester.AUTHMETHOD_DEVICEKEY,
	}

	t.Record, err = t.CreateDmpProduct(cp)
	t.Must(err)
	t.AssertTrue("创建LWM2M产品", t.Record.IsSuccess())
	pInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.ProductID = pInfo.Data.ID
	t.Params.MasterKey = pInfo.Data.MasterKey

	//添加设备
	imei := random.RandString()
	imsi := "test"
	lwDev := tester.CreateLWM2MDevicePl{
		Name:         "lwm2m_device",
		IMEI:         imei,
		IMSI:         imsi,
		Subscription: tester.LWM2M_SUBSCRIPTION_ENABLED,
	}
	t.Record, err = t.CreateLWM2MDevice(t.Params.ProductID, lwDev)
	t.Must(err)
	t.AssertTrue("添加LWM2M设备", t.Record.IsSuccess())

	dInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.DeviceID = dInfo.Data.ID
	t.Params.ApiKey = dInfo.Data.ApiKey

	t.Logger.Println("2秒后启动客户端")
	time.Sleep(2 * time.Second)

	//启动客户端
	nc := nbclient.NbiotClient{
		Addr: conf.Lwm2mServer,
		IMEI: imei,
		IMSI: imsi,
	}

	go func() {
		nc.Boot()
	}()

	secs := 10
	t.Logger.Printf("给客户端%v秒时间上行数据", secs)
	time.Sleep(time.Duration(secs) * time.Second)

	t.Logger.Println("开始下发命令...")

	// 读命令
	readCmd := tester.IssueLWM2MCommandPl{
		DeviceId:     t.Params.DeviceID,
		FunctionType: tester.LWM2M_PROPERTY_GET,
		Identifier:   "_3303_0_5700",
	}
	t.Record, err = t.IssueLWM2MCommand(readCmd)
	t.Must(err)
	readDp, ok := t.Record.Response.(*tester.CmdResp).Data.Value.(float64)
	if !ok {
		t.FailTest(t.Record.GetMsg())
		t.FailTest(t.Record.ReqBody)
	} else {
		t.AssertFloat64("读命令可以获取到数据点", readDp, 23.5)
	}

	// 对没有写命令权限的功能下发写命令
	errCmd1 := tester.IssueLWM2MCommandPl{
		DeviceId:        t.Params.DeviceID,
		FunctionType:    tester.LWM2M_PROPERTY_SET,
		Identifier:      "_3303_0_5700",
		IdentifierValue: 77.7,
	}
	t.Record, err = t.IssueLWM2MCommand(errCmd1)
	t.Must(err)
	t.AssertStringContains("对没有写命令权限的功能下发写命令后错误提示语正确", t.Record.GetMsg(), "没有写权限")

	// 对没有执行命令权限的功能下发执行命令
	errCmd3 := tester.IssueLWM2MCommandPl{
		DeviceId:        t.Params.DeviceID,
		FunctionType:    tester.LWM2M_PROPERTY_EXECUTE,
		Identifier:      "_3303_0_5700",
		IdentifierValue: 55.5,
	}
	t.Record, err = t.IssueLWM2MCommand(errCmd3)
	t.Must(err)
	t.AssertStringContains("对没有执行命令权限的功能下发执行命令错误提示语应正确", t.Record.GetMsg(), "没有执行权限")

	// 写命令
	writeCmd := tester.IssueLWM2MCommandPl{
		DeviceId:        t.Params.DeviceID,
		FunctionType:    tester.LWM2M_PROPERTY_SET,
		Identifier:      "_3308_0_5900",
		IdentifierValue: 12345,
	}
	t.Record, err = t.IssueLWM2MCommand(writeCmd)
	t.Must(err)
	t.AssertTrue("写命令可以成功下发", t.Record.IsSuccess())

	// 执行命令
	execCmd := tester.IssueLWM2MCommandPl{
		DeviceId:        t.Params.DeviceID,
		FunctionType:    tester.LWM2M_PROPERTY_EXECUTE,
		Identifier:      "_3303_0_5605",
		IdentifierValue: []byte{23, 23, 23, 23, 23},
	}
	t.Record, err = t.IssueLWM2MCommand(execCmd)
	t.Must(err)
	t.AssertTrue("执行命令可以成功下发", t.Record.IsSuccess())

	// 对没有读命令权限的功能下发读命令
	errCmd2 := tester.IssueLWM2MCommandPl{
		DeviceId:     t.Params.DeviceID,
		FunctionType: tester.LWM2M_PROPERTY_GET,
		Identifier:   "_3303_0_5605",
	}
	t.Record, err = t.IssueLWM2MCommand(errCmd2)
	t.Must(err)
	t.AssertStringContains("对没有读命令权限的功能下发读命令错误提示语应正确", t.Record.GetMsg(), "没有读权限")

	// 查询数据点1
	t.Record, err = t.GetMultipleLatestDataPoints(t.Params.DeviceID, []string{"SensorValue_3303_0_5700"})
	t.Must(err)
	if d1 := t.Record.Response.(*tester.DataPointResp).Data; len(d1) > 0 {
		t.AssertFloat64("可以获取到Temperature数据点", d1[0].Value.(float64), 23.5)
	} else {
		t.FailTest("未获取到Temperature数据点，测试失败")
	}

	// 查询数据点2
	t.Record, err = t.GetMultipleLatestDataPoints(t.Params.DeviceID, []string{"SetPointValue_3308_0_5900"})
	t.Must(err)

	if d2 := t.Record.Response.(*tester.DataPointResp).Data; len(d2) > 0 {
		t.AssertFloat64("可以获取到SetPoint数据点", d2[0].Value.(float64), 77.777)
	} else {
		t.FailTest("未获取到SetPoint数据点")
	}

	if onHold == true {
		wg.Add(1)
		wg.Wait()
	}
}

func tearDown(t *tester.Tester, resultChan chan tester.TestResult) {
	//删除产品
	var err error
	t.Record, err = t.DeleteDmpProduct(t.Params.ProductID)
	t.Must(err)
	t.AssertTrue("删除LWM2M设备", t.Record.IsSuccess())

	//判断套件是否成功并打印结果
	t.DisplayResult(resultChan)
}
