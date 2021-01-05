package ruleengine

import (
	"encoding/json"
	"fmt"
	"oes-acceptance-testing/v1/internal/tester"
	"oes-acceptance-testing/v1/pkg/helpers/conf"
	"oes-acceptance-testing/v1/pkg/helpers/random"
	"oes-acceptance-testing/v1/pkg/mqttClient"
	"os"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func NewRuleEngineMqttHandler(triggerChan chan int, setReply string) MQTT.MessageHandler {
	return func(c MQTT.Client, msg MQTT.Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())

		var typeCheck tester.Command
		if err := json.Unmarshal(msg.Payload(), &typeCheck); err != nil {
			fmt.Printf("NewMqttHandler: Error while unmarshaling command: %v\n", err)
			return
		}

		// use the Command struct to judge method
		switch typeCheck.FunctionType {

		case "propertySet": // "set" means this is a write command
			// so we send back its uuid as response, we will not handle the value sent by the command due ot its not part of the test
			var cmd tester.SetCommand
			if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
				fmt.Printf("NewMqttHandler: Error while unmarshaling command: %v\n", err)
				return
			}
			resp := []byte(fmt.Sprintf(`{"id": "%v", "code": 0, "msg": "Respond to set command, Kupo!"}`, cmd.Id))

			fmt.Println("回复Topic:", setReply)
			fmt.Println("回复: " + string(resp))
			if token := c.Publish(setReply, byte(0), false, resp); token.Wait() && token.Error() != nil {
				fmt.Println(token.Error())
				os.Exit(1)
			}
			triggerChan <- 1
			fmt.Println("notified trigger chan!")
		}
	}
}

func bootMqttClient(t *tester.Tester, triggerChan chan int) (client MQTT.Client, pid, did int) {
	var err error
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
	pt1 := tester.CreateDmpPropertyPl{
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
	t.Record, err = t.CreateDmpProperty(t.Params.ProductID, pt1)
	t.Must(err)
	t.AssertTrue("添加功能someDouble", t.Record.IsSuccess())

	// 添加功能
	pt2 := tester.CreateDmpPropertyPl{
		Name:       "someBoolean",
		Identifier: "someBoolean",
		AccessMode: 2,
		Type:       tester.DATATYPE_BOOLEAN,
		Unit:       "some_unit",
		Special: tester.Special{
			EnumArrays: []tester.EnumArray{
				{
					Key:      0,
					Describe: "false",
				},
				{
					Key:      1,
					Describe: "true",
				},
			},
		},
	}
	t.Record, err = t.CreateDmpProperty(t.Params.ProductID, pt2)
	t.Must(err)
	t.AssertTrue("添加功能someBoolean", t.Record.IsSuccess())

	//添加设备
	cd := tester.CreateDmpDevicePl{
		Name:        "rule_device",
		Description: "some device",
	}

	t.Record, err = t.CreateDmpDevice(t.Params.ProductID, cd)
	t.Must(err)
	t.AssertTrue("添加设备", t.Record.IsSuccess())
	dInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.DeviceID = dInfo.Data.ID
	t.Params.ApiKey = dInfo.Data.ApiKey

	// 生成新的MQTT客户端，其中自带处理命令用的handler
	setReplyTopic := fmt.Sprintf("$sys/%v/%v/thing/property/set_reply", t.Params.ProductID, t.Params.DeviceID)
	setTopic := fmt.Sprintf("$sys/%v/%v/thing/property/set", t.Params.ProductID, t.Params.DeviceID)

	c := mqttClient.EdgeMqttClient{
		ServerHostAndPort: conf.MqttBroker,
		ProductID:         strconv.Itoa(t.Params.ProductID),
		DeviceKey:         t.Params.ApiKey,
		DeviceName:        "rule_mqtt_device",
		DeviceID:          strconv.Itoa(t.Params.DeviceID),
		AuthMsg: mqttClient.AuthMessage{
			SasToken: mqttClient.SasToken{
				Version: "2018-10-31",
				Et:      time.Now().AddDate(0, 3, 0).Unix(),
				Method:  mqttClient.AUTH_SHA1,
			},
		},
	}

	c.NewMqttClient(600, NewRuleEngineMqttHandler(triggerChan, setReplyTopic))

	// 设备鉴权+上线
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		t.Logger.Fatalf("MQTT客户端连接错误：%v", token.Error())
	}

	t.Logger.Println("连接成功！等待2秒...")
	time.Sleep(2 * time.Second)

	// 订阅命令Topic
	if token := c.Subscribe(setTopic, 0, nil); token.Wait() && token.Error() != nil {
		t.Logger.Fatalf("MQTT客户端订阅错误：%v", token.Error())
	}
	return c, t.Params.ProductID, t.Params.DeviceID
}
