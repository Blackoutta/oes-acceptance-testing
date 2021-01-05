package httpscene

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"oes-acceptance-testing/v1/internal/tester"
	"oes-acceptance-testing/v1/pkg/helpers/conf"
	"oes-acceptance-testing/v1/pkg/helpers/logger"
	"oes-acceptance-testing/v1/pkg/helpers/random"
	"oes-acceptance-testing/v1/pkg/httpClient"
	"sync"
	"time"
)

var wg sync.WaitGroup

const (
	suiteName = "HTTP测试套件"
	td        = true
)

func RunHTTPTest(resultChan chan tester.TestResult, onHold bool) {
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

	// 创建HTTP产品
	cp := tester.CreateDmpProductPl{
		Name:                 "http_" + random.RandString(),
		Description:          "this is an http protocol product",
		ProtocolType:         tester.PROTOCOL_HTTP,
		NodeType:             tester.NODETYPE_DEVICE,
		Model:                tester.MODEL_OBJECTMODEL,
		AuthenticationMethod: tester.AUTHMETHOD_DEVICEKEY,
		NetworkMethod:        1,
	}

	t.Record, err = t.CreateDmpProduct(cp)
	t.Must(err)
	t.AssertTrue("创建HTTP产品", t.Record.IsSuccess())
	pInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.ProductID = pInfo.Data.ID
	t.Params.MasterKey = pInfo.Data.MasterKey

	// 添加功能
	cpr := tester.CreateDmpPropertyPl{
		Name:       "someDouble",
		Identifier: "someDouble",
		AccessMode: 2,
		Type:       tester.DATATYPE_FLOAT64,
		Unit:       "some_unit",
		Minimum:    float64(-999999.999999999),
		Maximum:    float64(999999.999999999),
		Special: tester.Special{
			Step: 0.000000001,
		},
	}
	t.Record, err = t.CreateDmpProperty(t.Params.ProductID, cpr)
	t.Must(err)
	t.AssertTrue("添加功能1", t.Record.IsSuccess())

	// 添加功能
	cpr2 := tester.CreateDmpPropertyPl{
		Name:       "someFloat",
		Identifier: "someFloat",
		AccessMode: 2,
		Type:       tester.DATATYPE_FLOAT64,
		Unit:       "some_unit",
		Minimum:    float64(-999.999),
		Maximum:    float64(999.999),
		Special: tester.Special{
			Step: 0.001,
		},
	}
	t.Record, err = t.CreateDmpProperty(t.Params.ProductID, cpr2)
	t.Must(err)
	t.AssertTrue("添加功能2", t.Record.IsSuccess())

	//添加设备
	cd := tester.CreateDmpDevicePl{
		Name:        "http_device",
		Description: "some device",
	}

	t.Record, err = t.CreateDmpDevice(t.Params.ProductID, cd)
	t.Must(err)
	t.AssertTrue("添加设备", t.Record.IsSuccess())
	dInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.DeviceID = dInfo.Data.ID
	t.Params.ApiKey = dInfo.Data.ApiKey

	// 准备上行请求
	oc := httpClient.NewOesHttpClient(t.Params.ProductID, t.Params.DeviceID, t.Params.ApiKey, conf.HttpServer)

	pl := `{"someDouble": 999999.999999999, "someFloat": 999.999}`
	r1 := oc.NewUpwardRequest(conf.HttpServer, pl)

	// 上行数据
	go func() {
		for i := 0; i < 3; i++ {
			resp, err := oc.Client.Do(r1)
			if err != nil {
				t.FailTest(fmt.Sprintf("error after sending upward request: %v", err))
				panic(err)
			}
			defer resp.Body.Close()
			t.Logger.Printf("data sent: %v", pl)
			if resp.StatusCode != 200 {
				err := fmt.Errorf("上行请求发送后响应码不为200，测试失败。")
				t.FailTest(err.Error())
				panic(err)
			}
			time.Sleep(time.Second)
		}
	}()

	time.Sleep(5 * time.Second)

	// GetMultipleLatestDataPoints
	t.Record, err = t.GetMultipleLatestDataPoints(t.Params.DeviceID, []string{"someDouble"})
	t.Must(err)
	t.AssertTrue("获取数据点", t.Record.IsSuccess())

	dp := t.Record.Response.(*tester.DataPointResp).Data
	if len(dp) < 1 {
		t.FailTest(fmt.Sprintf("未获取到数据点或数据点不全，获取到的数据点数量: %v\n", len(dp)))
		t.Logger.Println(t.Record.ResponseString)
	} else {
		t.Logger.Println(t.Record.ResponseString)
		t.AssertFloat64("平台入库的数据点应与上行的数据点一致(double)", 999999.999999999, dp[0].Value.(float64))
	}

	// 检查上行数据点应已入库2
	t.Record, err = t.GetMultipleLatestDataPoints(t.Params.DeviceID, []string{"someFloat"})
	t.Must(err)
	t.AssertTrue("获取数据点", t.Record.IsSuccess())

	dp2 := t.Record.Response.(*tester.DataPointResp).Data
	if len(dp) < 1 {
		t.FailTest(fmt.Sprintf("未获取到数据点或数据点不全，获取到的数据点数量: %v\n", len(dp)))
		t.Logger.Println(t.Record.ResponseString)
	} else {
		t.Logger.Println(t.Record.ResponseString)
		t.AssertFloat64("平台入库的数据点应与上行的数据点一致(float)", 999.999, dp2[0].Value.(float64))
	}

	time.Sleep(10 * time.Second)

	// 检查上行数据点数量
	// 检查上行数据点应入库(分页查询资源某时间段数据点)
	hour, err := time.ParseDuration("168h")
	hm := hour.Milliseconds()
	if err != nil {
		err := fmt.Errorf("error while parsing duration: %v", err)
		panic(err)
	}
	start := time.Now().Unix()*1000 - hm
	end := time.Now().Unix() * 1000
	t.Record, err = t.GetDataPointsInTimePeriod(t.Params.DeviceID, "someDouble", start, end, 1, 20)
	t.Must(err)
	data3 := t.Record.Response.(*tester.ContentDataPointResp).Data.Content
	t.AssertIntegerGreaterThan("检查上行数据点应入库(分页查询资源某时间段数据点)", 2, len(data3))

	// 上行超出物模型范围的数据应收到错误提示
	pl = `{"someDouble": -999999.9999999999}`
	r2 := oc.NewUpwardRequest(conf.HttpServer, pl)
	resp, err := oc.Client.Do(r2)
	t.Must(err)
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	t.AssertStringContains("上行超出物模型范围的数据应收到错误提示", string(bs), "属性值范围不正确")

	// 上行非物模型定义的数据类型应收到错误提示
	pl = `{"someDouble": "hahaha"}`
	r3 := oc.NewUpwardRequest(conf.HttpServer, pl)
	resp, err = oc.Client.Do(r3)
	t.Must(err)
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	t.AssertStringContains("上行非物模型定义的数据类型应收到错误提示", string(bs), "属性值类型错误")
	// 手动测试模式开关
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
	t.AssertTrue("删除HTTP设备", t.Record.IsSuccess())

	//判断套件是否成功并打印结果
	t.DisplayResult(resultChan)
}

func encryptPassword(pid, did, apiKey string) string {
	query := did + "&" + pid

	h := hmac.New(sha1.New, []byte(apiKey))
	h.Write([]byte(query))

	token := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return token
}
