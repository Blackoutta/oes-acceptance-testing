package mqttps

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"oes-acceptance-testing/v1/internal/tester"
	"oes-acceptance-testing/v1/pkg/helpers/conf"
	"oes-acceptance-testing/v1/pkg/helpers/logger"
	"oes-acceptance-testing/v1/pkg/helpers/random"
	"oes-acceptance-testing/v1/pkg/mqttClient"
	"strconv"
	"sync"
	"time"
)

var wg sync.WaitGroup

const (
	suiteName = "MQTT透传测试套件"
	td        = true
)

func RunMQTTPSTest(resultChan chan tester.TestResult, onHold bool) {
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

	// 创建MQTT产品
	cp := tester.CreateDmpProductPl{
		Name:                 "mqttps_" + random.RandString(),
		Description:          "some description",
		ProtocolType:         1,
		NodeType:             1,
		Model:                1,
		DataFormat:           2,
		AuthenticationMethod: 1,
		NetworkMethod:        1,
	}

	t.Record, err = t.CreateDmpProduct(cp)
	t.Must(err)
	t.AssertTrue("创建MQTT产品", t.Record.IsSuccess())
	pInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.ProductID = pInfo.Data.ID
	t.Params.MasterKey = pInfo.Data.MasterKey

	//添加功能
	t.CreateProperties()

	// 添加MQTT透传解析脚本
	bs, err := ioutil.ReadFile("data/mqttpsscript")
	if err != nil {
		panic(err)
	}

	script := tester.CreateTcpScriptPl{
		Content:    string(bs),
		ScriptType: "groovy",
	}
	t.Record, err = t.CreateTcpScript(t.Params.ProductID, t.Params.MasterKey, script)
	t.Must(err)
	t.AssertTrue("加入解析脚本", t.Record.IsSuccess())

	//添加设备
	cd := tester.CreateDmpDevicePl{
		Name:        "mqttps_device",
		Description: "some device",
	}

	t.Record, err = t.CreateDmpDevice(t.Params.ProductID, cd)
	t.Must(err)
	t.AssertTrue("添加设备", t.Record.IsSuccess())
	dInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.DeviceID = dInfo.Data.ID
	t.Params.ApiKey = dInfo.Data.ApiKey

	// Topic
	customizedPostTopic := fmt.Sprintf("customized/data/pid/%v/devkey/%v", t.Params.ProductID, t.Params.DeviceID)
	customizedCmdTopic := fmt.Sprintf("customized/cmd/pid/%v/devkey/%v", t.Params.ProductID, t.Params.DeviceID)
	customizedRespTopic := fmt.Sprintf("customized/resp/pid/%v/devkey/%v", t.Params.ProductID, t.Params.DeviceID)

	// 生成新的MQTT客户端，其中自带处理命令用的handler
	c := mqttClient.EdgeMqttClient{
		ServerHostAndPort: conf.MqttBroker,
		ProductID:         strconv.Itoa(t.Params.ProductID),
		DeviceKey:         t.Params.ApiKey,
		DeviceName:        "some_device",
		DeviceID:          strconv.Itoa(t.Params.DeviceID),
		AuthMsg: mqttClient.AuthMessage{
			SasToken: mqttClient.SasToken{
				Version: "2018-10-31",
				Et:      time.Now().AddDate(0, 3, 0).Unix(),
				Method:  mqttClient.AUTH_SHA1,
			},
		},
	}

	if err := c.NewMqttClient(600, mqttClient.NewMqttHandler(c.ProductID, c.DeviceID, customizedRespTopic)); err != nil {
		panic(err)
	}

	// 设备鉴权+上线
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		t.Logger.Fatalf("MQTT客户端连接错误：%v", token.Error())
	}

	t.Logger.Println("连接成功！等待2秒...")
	time.Sleep(2 * time.Second)

	subTopics := make(map[string]byte)
	subTopics[customizedCmdTopic] = 1

	if token := c.SubscribeMultiple(subTopics, nil); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	rawPl := tester.NewDataModel()

	pls := struct {
		Id      string                      `json:"id"`
		Version string                      `json:"version"`
		Params  map[string]tester.Datapoint `json:"params"`
	}{
		Id:      "123123123",
		Version: "1.0",
		Params:  rawPl,
	}
	pl, err := json.Marshal(pls)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(pl))

	// 上行数据
	qos := byte(1)
	retained := false

	go func() {
		for {
			if token := c.Publish(customizedPostTopic, qos, retained, pl); token.Wait() && token.Error() != nil {
				t.Logger.Fatalf("MQTT客户端Pub错误：%v", token.Error())
			}
			time.Sleep(5 * time.Second)
			fmt.Printf("已上行数据: %v\n", string(pl))
			fmt.Printf("已上行数据bytes: %v\n", []byte(pl))
		}
	}()

	time.Sleep(10 * time.Second)

	// 下发读命令
	mqttGet := tester.IssueMQTTCommandPl{
		DeviceId:     t.Params.DeviceID,
		FunctionType: tester.PROPERTY_GET,
		Identifier:   "someDouble",
	}

	t.Record, err = t.IssueMQTTCommand(mqttGet)
	t.Must(err)
	fmt.Println(t.Record.ResponseString)
	t.AssertStringContains("下发在线同步读命令应成功", t.Record.ResponseString, "999999.999999999")

	// 下发写命令
	mqttSet := tester.IssueMQTTCommandPl{
		DeviceId:        t.Params.DeviceID,
		FunctionType:    tester.PROPERTY_SET,
		Identifier:      "someDouble",
		IdentifierValue: 123.123,
	}
	t.Record, err = t.IssueMQTTCommand(mqttSet)
	t.Must(err)
	t.AssertTrue("下发在线同步写命令", t.Record.IsSuccess())

	// 检查上行数据点应已入库
	t.Record, err = t.GetMultipleLatestDataPoints(t.Params.DeviceID, []string{"someDouble"})
	t.Must(err)
	dp := t.Record.Response.(*tester.DataPointResp)
	t.AssertTrue("调用查询设备的多个属性值的最新值接口应成功", t.Record.IsSuccess())
	if len(dp.Data) < 1 {
		t.FailTest("获取最新数据点接口未获取到数据点!测试失败")
	} else {
		t.AssertFloat64("获取最新数据点的值应正确", rawPl["someDouble"].Value.(float64), dp.Data[0].Value.(float64))
	}

	// 检查上行数据点应入库(查询设备的多个属性值的最新值接口)
	t.Record, err = t.GetMultipleLatestDataPoints(t.Params.DeviceID, []string{"someDouble"})
	t.Must(err)
	data2 := t.Record.Response.(*tester.DataPointResp).Data
	t.AssertIntegerGreaterThan("检查上行数据点应入库(查询设备的多个属性值的最新值接口)", 0, len(data2))

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
	t.AssertIntegerGreaterThan("检查上行数据点应入库(分页查询资源某时间段数据点)", 0, len(data3))

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
	t.AssertTrue("删除MQTT设备", t.Record.IsSuccess())

	//判断套件是否成功并打印结果
	t.DisplayResult(resultChan)
}

// makePayload converts a json string to a payload that the edge platform can parse
func makePayload(jsonString string) []byte {
	j := jsonString
	jl := len(j)
	jb := make([]byte, 2)
	binary.BigEndian.PutUint16(jb, uint16(jl))
	payload := []byte{3}
	payload = append(payload, jb...)
	payload = append(payload, []byte(j)...)
	return payload
}
