package conf

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var (
	BaseURL               string
	AccessKeyID           string
	Secret                string
	TcpServer             string
	MqttBroker            string
	MqttDynamicRegister   string
	MqttsBroker           string
	Lwm2mServer           string
	ModbusRTUServer       string
	HttpServer            string
	DataRouterDestination string
)

func init() {
	cm := make(map[string]string)

	configFile := os.Getenv("CONFIGFILE")
	if configFile == "" {
		log.Fatalln("环境变量 CONFIGFILE 读取失败，请正确设置 CONFIGFILE。例: export CONFIGFILE=config-test.json")
	}

	bs, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("error while reading config file: %v", err)
	}
	if err := json.Unmarshal(bs, &cm); err != nil {
		log.Fatalf("error while unmarshaling config file: %v", err)
	}

	BaseURL = cm["baseURL"]
	AccessKeyID = cm["accessKeyId"]
	Secret = cm["secret"]
	TcpServer = cm["tcpServer"]
	MqttBroker = cm["mqttBroker"]
	MqttDynamicRegister = cm["mqttDynamicRegister"]
	MqttsBroker = cm["mqttsBroker"]
	Lwm2mServer = cm["lwm2mServer"]
	ModbusRTUServer = cm["modbusRTUServer"]
	HttpServer = cm["httpServer"]
	DataRouterDestination = cm["dataRouterDestination"]
}
