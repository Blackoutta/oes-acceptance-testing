package mqttClient

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type command struct {
	FunctionType string `json:"functionType"`
	UUID         string `json:"uuid"`
	Identifier   string `json:"identifier"`
}

type setCommand struct {
	Id           string                 `json:"id"`
	Version      string                 `json:"version"`
	ProductId    int                    `json:"productId"`
	DeviceId     int                    `json:"deviceId"`
	FunctionType string                 `json:"functionType"`
	Params       map[string]interface{} `json:"params"`
	Timeout      int                    `json:"timeout"`
	Mode         int                    `json:"mode"`
}

type Check struct {
	ID  int
	Msg string
}

func printMsg(msg MQTT.Message) {
	log.Printf("\n收到下行...\n")
	log.Println("↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓")
	log.Printf("TOPIC: %s\n", msg.Topic())
	log.Printf("MSG: %s\n", msg.Payload())
	log.Println("↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓")
	log.Println()
}

func NewShadowUpdateRespHandler(check chan Check) MQTT.MessageHandler {
	return func(c MQTT.Client, msg MQTT.Message) {
		printMsg(msg)
		checkRequest := Check{
			ID:  3,
			Msg: string(msg.Payload()),
		}
		check <- checkRequest
	}
}

func NewShadowGetRespHandler(check chan Check) MQTT.MessageHandler {
	return func(c MQTT.Client, msg MQTT.Message) {
		printMsg(msg)
		checkRequest := Check{
			ID:  2,
			Msg: string(msg.Payload()),
		}
		check <- checkRequest
	}
}

func NewNTPHandler(check chan Check) MQTT.MessageHandler {
	return func(c MQTT.Client, msg MQTT.Message) {
		printMsg(msg)
		// 设备端收到服务端的时间记为${deviceRecvTime}，则设备上的精确时间为：(${serverRecvTime} + ${serverSendTime} + ${deviceRecvTime} - ${deviceSendTime}) / 2
		npt := struct {
			DeviceSendTime int64
			ServerRecvTime int64
			ServerSendTime int64
		}{}
		if err := json.Unmarshal(msg.Payload(), &npt); err != nil {
			panic(err)
		}
		exactTime := (npt.ServerRecvTime + npt.ServerSendTime + time.Now().Unix()*1000 - npt.DeviceSendTime) / 2
		fmt.Printf("当前精确Unix时间为: %v, 换算为可读时间为：%s\n", exactTime, time.Unix(exactTime/1000, 0))
		checkRequest := Check{
			ID:  1,
			Msg: fmt.Sprintf("%d", exactTime),
		}
		check <- checkRequest
	}
}

func NewMqttHandler(productId, deviceId string, setReply string) MQTT.MessageHandler {
	return func(c MQTT.Client, msg MQTT.Message) {
		printMsg(msg)

		var typeCheck command
		if err := json.Unmarshal(msg.Payload(), &typeCheck); err != nil {
			fmt.Printf("NewMqttHandler: Error while unmarshaling command: %v\n", err)
			return
		}

		// use the Command struct to judge method
		switch typeCheck.FunctionType {

		case "propertySet": // "set" means this is a write command
			// so we send back its uuid as response, we will not handle the value sent by the command due ot its not part of the test
			var cmd setCommand
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
		}
	}
}
