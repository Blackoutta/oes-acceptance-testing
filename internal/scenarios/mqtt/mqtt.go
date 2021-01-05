package mqtt

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
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
	suiteName = "MQTT测试套件"
	td        = true
)

func RunMQTTTest(resultChan chan tester.TestResult, onHold bool) {
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
		Name:                 "mqtt_" + random.RandString(),
		Description:          "some description",
		ProtocolType:         1,
		NodeType:             1,
		Model:                1,
		DataFormat:           1,
		AuthenticationMethod: 1,
		DynamicRegister:      1,
		NetworkMethod:        1,
	}

	t.Record, err = t.CreateDmpProduct(cp)
	fmt.Println(t.Record.ReqBody)
	t.Must(err)
	t.AssertTrue("创建MQTT产品", t.Record.IsSuccess())
	pInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.ProductID = pInfo.Data.ID
	t.Params.MasterKey = pInfo.Data.MasterKey
	fmt.Printf("masterkey: %v\n", t.Params.MasterKey)

	//添加功能

	t.CreateProperties()

	//添加设备
	cd := tester.CreateDmpDevicePl{
		Name:        "mqtt_device",
		Description: "this is a mqtt device",
	}

	t.Record, err = t.CreateDmpDevice(t.Params.ProductID, cd)
	t.Must(err)
	t.AssertTrue("添加设备", t.Record.IsSuccess())
	dInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.DeviceID = dInfo.Data.ID
	t.Params.ApiKey = dInfo.Data.ApiKey

	// Topic
	setTopic := fmt.Sprintf("$sys/%v/%v/thing/property/set", t.Params.ProductID, t.Params.DeviceID)
	setReplyTopic := fmt.Sprintf("$sys/%v/%v/thing/property/set_reply", t.Params.ProductID, t.Params.DeviceID)
	postReplyTopic := fmt.Sprintf("$sys/%v/%v/thing/property/post/reply", t.Params.ProductID, t.Params.DeviceID)
	postTopic := fmt.Sprintf("$sys/%v/%v/thing/property/post", t.Params.ProductID, t.Params.DeviceID)
	shadowUpdateRespTopic := fmt.Sprintf("shadow/update/resp/pid/%v/devkey/%v", t.Params.ProductID, t.Params.DeviceID)
	shadowGetResp := fmt.Sprintf("shadow/get/resp/pid/%v/devkey/%v", t.Params.ProductID, t.Params.DeviceID)
	configTopic := fmt.Sprintf("$sys/%v/%v/thing/config/push", t.Params.ProductID, t.Params.DeviceID)
	ntpRespTopic := fmt.Sprintf("/ext/ntp/%v/%v/response", t.Params.ProductID, t.Params.DeviceID)

	fmt.Println("开始动态注册...")

	// // 进行动态注册
	// dynamicRegisterURL := fmt.Sprintf("http://%v/dynamicregister", conf.MqttDynamicRegister)
	// fmt.Println("注册URL:", dynamicRegisterURL)
	// registerClient := &http.Client{
	// 	Timeout: 10 * time.Second,
	// }

	// drToken, err := DynamicRegisterToken(t.Params.ProductID, t.Params.MasterKey, "mqtt_device")
	// if err != nil {
	// 	fmt.Printf("生成动态注册token时发生错误: %v\n", err)
	// }

	// dr := fmt.Sprintf(`{"deviceName": "%v", "pid": "%v", "token": "%v"}`, "mqtt_device", t.Params.ProductID, drToken)
	// fmt.Println("注册body:" + dr)

	// registerResp, err := registerClient.Post(dynamicRegisterURL, "application/json", bytes.NewReader([]byte(dr)))
	// if err != nil {
	// 	fmt.Printf("请求产品级鉴权时发生错误: %v\n", err)
	// }

	// registerRespBody := make([]byte, 0)
	// if registerResp != nil {
	// 	registerRespBody, err = ioutil.ReadAll(registerResp.Body)
	// 	if err != nil {
	// 		fmt.Printf("读取动态注册响应时发生错误: %v\n", err)
	// 	}

	// 	fmt.Printf("动态注册获取到设备信息: %v\n", string(registerRespBody))
	// 	registerResp.Body.Close()
	// }

	// drResp := struct {
	// 	Data struct {
	// 		DeviceId     int    `json:"deviceId"`
	// 		DeviceSecret string `json:"deviceSecret"`
	// 	} `json:"data"`
	// }{}
	// if err := json.Unmarshal(registerRespBody, &drResp); err != nil {
	// 	t.FailTest("动态注册失败...")
	// }

	// t.AssertInteger("动态注册获取到的deviceId与实际deviceId一致", t.Params.DeviceID, drResp.Data.DeviceId)
	// t.AssertStringEqual("动态注册获取到的device secret与实际device secret一致", t.Params.ApiKey, drResp.Data.DeviceSecret)

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

	if err := c.NewMqttClient(600, mqttClient.NewMqttHandler(c.ProductID, c.DeviceID, setReplyTopic)); err != nil {
		panic(err)
	}

	// 设备鉴权+上线
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		t.Logger.Fatalf("MQTT客户端连接错误：%v", token.Error())
	}

	t.Logger.Println("连接成功！等待2秒...")
	time.Sleep(2 * time.Second)

	subTopics := make(map[string]byte)
	subTopics[setTopic] = 1
	subTopics[postReplyTopic] = 1
	subTopics[configTopic] = 1

	if token := c.SubscribeMultiple(subTopics, nil); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	rawPl := tester.NewDataModel()

	pls := struct {
		Params map[string]tester.Datapoint `json:"params"`
	}{
		Params: rawPl,
	}
	pl, err := json.Marshal(pls)
	if err != nil {
		panic(err)
	}

	// 上行数据
	qos := byte(1)
	retained := false
	go func() {
		for {
			if token := c.Publish(postTopic, qos, retained, pl); token.Wait() && token.Error() != nil {
				t.Logger.Fatalf("MQTT客户端Pub错误：%v", token.Error())
			}
			fmt.Printf("已上行数据: %v\n", string(pl))
			// fmt.Printf("已上行数据bytes: %v\n", []byte(pl))
			time.Sleep(5 * time.Second)
		}
	}()

	time.Sleep(1 * time.Second)

	// 下发读命令
	for k, v := range rawPl {
		if k == "someBytes" {
			continue
		}
		mqttGet := tester.IssueMQTTCommandPl{
			DeviceId:     t.Params.DeviceID,
			FunctionType: tester.PROPERTY_GET,
			Identifier:   k,
		}

		t.Record, err = t.IssueMQTTCommand(mqttGet)
		t.Must(err)
		t.AssertStringContains(fmt.Sprintf("下发在线同步读命令应成功(%s)", k), t.Record.ResponseString, fmt.Sprintf("%v", v.Value))
	}

	// 下发写命令
	for k, v := range rawPl {
		if k == "someBytes" {
			continue
		}
		mqttSet := tester.IssueMQTTCommandPl{
			DeviceId:        t.Params.DeviceID,
			FunctionType:    tester.PROPERTY_SET,
			Identifier:      k,
			IdentifierValue: v.Value,
		}
		t.Record, err = t.IssueMQTTCommand(mqttSet)
		t.Must(err)
		t.AssertTrue(fmt.Sprintf("下发在线同步写命令应成功(%s)", k), t.Record.IsSuccess())
	}

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
	plist := make([]string, 0, 9)
	for k := range rawPl {
		plist = append(plist, k)
	}
	t.Record, err = t.GetMultipleLatestDataPoints(t.Params.DeviceID, plist)
	t.Must(err)
	data2 := t.Record.Response.(*tester.DataPointResp).Data
	t.AssertIntegerGreaterThan("检查上行数据点应入库, 且入库刷领大于7(查询设备的多个属性值的最新值接口)", 7, len(data2))

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

	checkChan := make(chan mqttClient.Check, 1)

	// 设备上行设备影子
	if token := c.Subscribe(shadowUpdateRespTopic, qos, mqttClient.NewShadowUpdateRespHandler(checkChan)); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("上行设备影子...")
	postShadowTopic := fmt.Sprintf("shadow/update/pid/%v/devkey/%v", c.ProductID, c.DeviceID)
	postShadowPl := `{"method": "update",
				  "state": {
					"reported": {
						"someString": "shadow string"
					}
				  },
				  "version": 1}`

	if token := c.Publish(postShadowTopic, qos, retained, postShadowPl); token.Wait() && token.Error() != nil {
		t.Logger.Fatalf("MQTT客户端Pub错误：%v", token.Error())
	}

	// 设备获取设备影子
	if token := c.Subscribe(shadowGetResp, qos, mqttClient.NewShadowGetRespHandler(checkChan)); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("获取设备影子...")
	getShadowPl := `{"method": "get"}`
	getShadowTopic := fmt.Sprintf("shadow/get/pid/%v/devkey/%v", c.ProductID, c.DeviceID)
	if token := c.Publish(getShadowTopic, qos, retained, getShadowPl); token.Wait() && token.Error() != nil {
		t.Logger.Fatalf("MQTT客户端Pub错误：%v", token.Error())
	}

	// 发送时钟同步请求
	if token := c.Subscribe(ntpRespTopic, qos, mqttClient.NewNTPHandler(checkChan)); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("发送时钟同步请求...")
	ntpPl := fmt.Sprintf(`{"deviceSendTime": %v}`, time.Now().Unix()*1000)
	ntpTopic := fmt.Sprintf(`/ext/ntp/%v/%v/request`, c.ProductID, c.DeviceID)
	if token := c.Publish(ntpTopic, qos, retained, ntpPl); token.Wait() && token.Error() != nil {
		t.Logger.Fatalf("MQTT客户端Pub错误：%v", token.Error())
	}

	// 检查
	var count int
Loop:
	for count < 3 {
		select {
		case c := <-checkChan:
			switch c.ID {
			case 1:
				fmt.Println("收到ntp响应...")
				now := fmt.Sprintf("%d", time.Now().Unix()*1000)
				t.AssertStringContains("应能获取到ntp服务回复的时间", c.Msg, now[:2])
				count++
			case 2:
				fmt.Println("收到设备获取设备影子响应...")
				t.AssertStringContains("设备获取影子应包含刚才上报的设备影子属性", c.Msg, "shadow string")
				count++
			case 3:
				fmt.Println("收到设备上行影子响应...")
				t.AssertStringContains("设备上行影子应成功", c.Msg, "success")
				count++
			}
		case <-time.After(15 * time.Second):
			t.FailTest("15秒未获取到检查项，检查超时....")
			break Loop
		}
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

func DynamicRegisterToken(pid int, pkey, deviceName string) (string, error) {
	target := strconv.Itoa(pid) + "&" + deviceName
	mac := hmac.New(sha1.New, []byte(pkey))
	_, err := mac.Write([]byte(target))
	if err != nil {
		return "", err
	}
	src := mac.Sum(nil)
	dst := base64.StdEncoding.EncodeToString(src)
	return string(dst), nil
}
