package datarouter

import (
	"database/sql"
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

	MQTT "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/go-sql-driver/mysql"
)

type consumerMessage struct {
	DeviceID string `json:"deviceID,omitempty"`
}

const (
	suiteName = "数据路由测试套件"
	td        = true
)

var wg sync.WaitGroup

func RunDataRouterTest(resultChan chan tester.TestResult, onHold bool) {
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
		NetworkMethod:        1,
	}

	t.Record, err = t.CreateDmpProduct(cp)
	t.Must(err)
	t.AssertTrue("创建MQTT产品", t.Record.IsSuccess())
	pInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.ProductID = pInfo.Data.ID
	t.Params.MasterKey = pInfo.Data.MasterKey

	// 添加功能
	t.CreateProperties()

	//添加设备
	cd := tester.CreateDmpDevicePl{
		Name:        "some_device",
		Description: "some device",
	}

	t.Record, err = t.CreateDmpDevice(t.Params.ProductID, cd)
	t.Must(err)
	t.AssertTrue("添加设备", t.Record.IsSuccess())
	dInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.DeviceID = dInfo.Data.ID
	t.Params.ApiKey = dInfo.Data.ApiKey

	// 生成新的OES MQTT客户端，其中自带处理命令用的handler
	// Topic
	setReplyTopic := fmt.Sprintf("$sys/%v/%v/thing/property/set_reply", t.Params.ProductID, t.Params.DeviceID)
	postTopic := fmt.Sprintf("$sys/%v/%v/thing/property/post", t.Params.ProductID, t.Params.DeviceID)

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

	// 上行数据
	go func() {
		for {
			if token := c.Publish(postTopic, byte(1), false, pl); token.Wait() && token.Error() != nil {
				t.Logger.Printf("MQTT客户端Pub错误：%v", token.Error())
				time.Sleep(3 * time.Second)
				continue
			}
			fmt.Println("已上行数据...")
			time.Sleep(3 * time.Second)
		}
	}()

	// ==============创建目的地===============//
	//rest
	restAddr := tester.CreateDMPRouterAddressPl{
		Name: "rest_addr_" + random.RandString(),
		Type: tester.ROUTER_ADDR_REST,
		Env:  1,
		Host: conf.DataRouterDestination,
		Port: 9005,
		Path: "/automation",
	}
	t.Record, err = t.CreateDMPRouterAddress(restAddr)
	t.Must(err)
	t.AssertTrue("创建rest地址", t.Record.IsSuccess())
	t.Params.RestAddrID = t.Record.Response.(*tester.SingleDataResp).Data

	//kafka
	kafkaAddr := tester.CreateDMPRouterAddressPl{
		Name:  "kafka_addr_" + random.RandString(),
		Type:  tester.ROUTER_ADDR_KAFKA,
		Env:   1,
		Addr:  conf.DataRouterDestination + ":9092",
		Topic: "edgetest_automation",
	}
	t.Record, err = t.CreateDMPRouterAddress(kafkaAddr)
	t.Must(err)
	t.AssertTrue("创建kafka地址", t.Record.IsSuccess())
	t.Params.KAFKAAddrID = t.Record.Response.(*tester.SingleDataResp).Data

	//mqtt
	mqttAddr := tester.CreateDMPRouterAddressPl{
		Name:     "mqtt_addr_" + random.RandString(),
		Type:     tester.ROUTER_ADDR_MQTT,
		Env:      1,
		Host:     conf.DataRouterDestination,
		Port:     18830,
		ClientId: "tester",
		Username: "cmiot",
		Password: "Iot@10086",
		Topic:    "/test/forward",
	}
	t.Record, err = t.CreateDMPRouterAddress(mqttAddr)
	t.Must(err)
	t.AssertTrue("创建mqtt地址", t.Record.IsSuccess())
	t.Params.MQTTAddrID = t.Record.Response.(*tester.SingleDataResp).Data

	//mysql
	mysqlAddr := tester.CreateDMPRouterAddressPl{
		Name:     "mysql_addr_" + random.RandString(),
		Type:     tester.ROUTER_ADDR_MYSQL,
		Env:      1,
		Host:     conf.DataRouterDestination,
		Port:     3306,
		Username: "root",
		Password: "Iot!@10086",
		Database: "edgetester",
		Table:    "test",
	}
	t.Record, err = t.CreateDMPRouterAddress(mysqlAddr)
	t.Must(err)
	t.AssertTrue("创建mysql地址", t.Record.IsSuccess())
	t.Params.MYSQLAddrID = t.Record.Response.(*tester.SingleDataResp).Data

	// ======================创建路由======================//
	//rest
	restRouter := tester.CreateDMPRouterPl{
		Name:            "rest_router_" + random.RandString(),
		Env:             1,
		Type:            tester.ROUTER_TYPE_STANDARD,
		AddressType:     tester.ROUTER_ADDR_REST,
		AddressConfigId: t.Params.RestAddrID,
		Format:          tester.ROUTER_FORMAT_JSON,
		Compression:     tester.ROUTER_COMPRESSION_NONE,
		Filter: tester.Filter{
			DevIdentifiers: []tester.DevIdentifier{
				{Pid: strconv.Itoa(t.Params.ProductID)},
			},
		},
	}
	t.Record, err = t.CreateDMPDataRouter(restRouter)
	t.Must(err)
	t.AssertTrue("创建rest路由", t.Record.IsSuccess())
	t.Params.RestRouterID = t.Record.Response.(*tester.SingleDataResp).Data

	//kafka
	kafkaRouter := tester.CreateDMPRouterPl{
		Name:            "kafka_router_" + random.RandString(),
		Env:             1,
		Type:            tester.ROUTER_TYPE_STANDARD,
		AddressType:     tester.ROUTER_ADDR_KAFKA,
		AddressConfigId: t.Params.KAFKAAddrID,
		Format:          tester.ROUTER_FORMAT_JSON,
		Compression:     tester.ROUTER_COMPRESSION_NONE,
		Filter: tester.Filter{
			DevIdentifiers: []tester.DevIdentifier{
				{Pid: strconv.Itoa(t.Params.ProductID)},
			},
		},
	}
	t.Record, err = t.CreateDMPDataRouter(kafkaRouter)
	t.Must(err)
	t.AssertTrue("创建kafka路由", t.Record.IsSuccess())
	t.Params.KafkaRouterID = t.Record.Response.(*tester.SingleDataResp).Data

	//mqtt
	mqttRouter := tester.CreateDMPRouterPl{
		Name:            "mqtt_router_" + random.RandString(),
		Env:             1,
		Type:            tester.ROUTER_TYPE_STANDARD,
		AddressType:     tester.ROUTER_ADDR_MQTT,
		AddressConfigId: t.Params.MQTTAddrID,
		Format:          tester.ROUTER_FORMAT_JSON,
		Compression:     tester.ROUTER_COMPRESSION_NONE,
		Filter: tester.Filter{
			DevIdentifiers: []tester.DevIdentifier{
				{Pid: strconv.Itoa(t.Params.ProductID)},
			},
		},
	}
	t.Record, err = t.CreateDMPDataRouter(mqttRouter)
	t.Must(err)
	t.AssertTrue("创建mqtt路由", t.Record.IsSuccess())
	t.Params.MQTTRouterID = t.Record.Response.(*tester.SingleDataResp).Data

	//mysql
	mysqlRouter := tester.CreateDMPRouterPl{
		Name:            "mysql_router_" + random.RandString(),
		Env:             1,
		Type:            tester.ROUTER_TYPE_STANDARD,
		AddressType:     tester.ROUTER_ADDR_MYSQL,
		AddressConfigId: t.Params.MYSQLAddrID,
		Format:          tester.ROUTER_FORMAT_JSON,
		Compression:     tester.ROUTER_COMPRESSION_NONE,
		Filter: tester.Filter{
			DevIdentifiers: []tester.DevIdentifier{
				{Pid: strconv.Itoa(t.Params.ProductID)},
			},
		},
	}
	t.Record, err = t.CreateDMPDataRouter(mysqlRouter)
	t.Must(err)
	t.AssertTrue("创建mysql路由", t.Record.IsSuccess())
	t.Params.MYSQLRouterID = t.Record.Response.(*tester.SingleDataResp).Data

	// =================启动路由=================//
	//rest
	t.Record, err = t.EnableDMPDataRouter(nil, t.Params.RestRouterID)
	t.Must(err)
	t.AssertTrue("启动rest路由", t.Record.IsSuccess())

	//mqtt
	t.Record, err = t.EnableDMPDataRouter(nil, t.Params.MQTTRouterID)
	t.Must(err)
	t.AssertTrue("启动mqtt路由", t.Record.IsSuccess())

	//mysql
	t.Record, err = t.EnableDMPDataRouter(nil, t.Params.MYSQLRouterID)
	t.Must(err)
	t.AssertTrue("启动mysql路由", t.Record.IsSuccess())

	//kafka
	t.Record, err = t.EnableDMPDataRouter(nil, t.Params.KafkaRouterID)
	t.Must(err)
	t.AssertTrue("启动kafka路由", t.Record.IsSuccess())

	// ============================验证数据转发应成功============================//

	//==============kafka consumer==============//
	t.Logger.Println("启动kafka消费者...")
	kafkaChan := make(chan []byte)
	kc, err := NewKafkaConsumer(fmt.Sprintf("%v:9092", conf.DataRouterDestination), "edgetest_automation")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			msg, err := kc.ReadMessage(-1)
			if err == nil {
				fmt.Println(string(msg.Value))
				kafkaChan <- msg.Value

			} else {
				// The client will automatically try to recover from all errors.
				fmt.Printf("Consumer error: %v (%v)\n", err, msg)
			}
		}
	}()

	//==============mqtt consumer==============//
	t.Logger.Println("启动mqtt消费者...")
	mqttChan := make(chan []byte, 1)
	mqttTopic := "/test/forward"
	var fh MQTT.MessageHandler = func(c MQTT.Client, msg MQTT.Message) {
		mqttChan <- msg.Payload()
	}

	mc := NewMQTTClient("consumer", "cmiot", "Iot@10086", conf.DataRouterDestination+":18830", fh)

	if token := mc.Connect(); token.Wait() && token.Error() != nil {
		panic(err)
	}
	t.Logger.Println("mqtt消费者连接成功")

	if token := mc.Subscribe(mqttTopic, 0, nil); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	//==============mysql consumer==============//
	t.Logger.Println("启动mysql消费者...")
	mysqlChan := make(chan string)
	go func() {
		db, err := sql.Open("mysql", fmt.Sprintf("root:Iot!@10086@tcp(%v:3306)/edgetester", conf.DataRouterDestination))
		if err != nil {
			panic(err)
		}
		defer db.Close()

		for i := 0; i < 5; i++ {
			rows, err := db.Query("SELECT device_id FROM test ORDER BY created DESC LIMIT 1")
			if err != nil {
				panic(err)
			}
			for rows.Next() {
				var deviceID string
				rows.Scan(&deviceID)
				mysqlChan <- deviceID
			}
			time.Sleep(3 * time.Second)
		}
	}()

	//=============rest consumer============//
	t.Logger.Println("启动rest消费者...")
	restChan := make(chan []byte)
	go func() {
		for i := 0; i < 5; i++ {
			resp, err := http.Get(fmt.Sprintf("http://%v:9005/getdata", conf.DataRouterDestination))
			if err != nil {
				panic(err)
			}
			bs, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			restChan <- bs
			time.Sleep(3 * time.Second)
		}
	}()

	//===================让消费者们消费15秒后，禁用所有路由=======================//
	go func() {
		t.Logger.Println("15秒后禁用所有路由")
		time.Sleep(15 * time.Second)
		//rest
		t.Record, err = t.DisableDMPDataRouter(nil, t.Params.RestRouterID)
		t.Must(err)
		t.AssertTrue("禁用rest路由", t.Record.IsSuccess())

		//mqtt
		t.Record, err = t.DisableDMPDataRouter(nil, t.Params.MQTTRouterID)
		t.Must(err)
		t.AssertTrue("禁用mqtt路由", t.Record.IsSuccess())

		//mysql
		t.Record, err = t.DisableDMPDataRouter(nil, t.Params.MYSQLRouterID)
		t.Must(err)
		t.AssertTrue("禁用mysql路由", t.Record.IsSuccess())

		//kafka
		t.Record, err = t.DisableDMPDataRouter(nil, t.Params.KafkaRouterID)
		t.Must(err)
		t.AssertTrue("禁用kafka路由", t.Record.IsSuccess())
	}()

	// 从各个转发服务器消费数据
	t.Logger.Println("开始消费...")
	var kafkaFinalDeviceID string
	var mqttFinalDeviceID string
	var mysqlFinalDeviceID string
	var restFinalDeviceID string
	var done bool
	for !done {
		select {
		case kafkaMsg := <-kafkaChan:
			var km consumerMessage
			if err := json.Unmarshal(kafkaMsg, &km); err != nil {
				t.FailTest("abnormal kafka message" + string(kafkaMsg))
			}
			kafkaFinalDeviceID = km.DeviceID
			t.Logger.Println("kafka msg deviceID: " + kafkaFinalDeviceID)
		case mqttMsg := <-mqttChan:
			var mm consumerMessage
			if err := json.Unmarshal(mqttMsg, &mm); err != nil {
				t.FailTest("abnormal mqtt message" + string(mqttMsg))
				panic(err)
			}
			mqttFinalDeviceID = mm.DeviceID
			t.Logger.Println("mqtt msg deviceID: " + mqttFinalDeviceID)
		case mysqlMsg := <-mysqlChan:
			mysqlFinalDeviceID = mysqlMsg
			t.Logger.Println("mysql msg deviceID: " + mysqlFinalDeviceID)
		case restMsg := <-restChan:
			var rm consumerMessage
			if err := json.Unmarshal(restMsg, &rm); err != nil {
				t.FailTest("rest message" + string(restMsg))
			}
			restFinalDeviceID = rm.DeviceID
			t.Logger.Println("rest msg deviceID: " + restFinalDeviceID)
		case <-time.After(20 * time.Second):
			t.Logger.Println("10秒未接收到消费信息，跳出消费循环，开始执行断言")
			done = true
		}
	}
	t.AssertStringEqual("rest转发应成功", strconv.Itoa(t.Params.DeviceID), restFinalDeviceID)
	t.AssertStringEqual("mqtt转发应成功", strconv.Itoa(t.Params.DeviceID), mqttFinalDeviceID)
	t.AssertStringEqual("mysql转发应成功", strconv.Itoa(t.Params.DeviceID), mysqlFinalDeviceID)
	t.AssertStringEqual("kafka转发应成功", strconv.Itoa(t.Params.DeviceID), kafkaFinalDeviceID)
}
