package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"oes-acceptance-testing/v1/pkg/mqttClient"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var mqttpsDesc = []string{
	"mqtt透传设备模拟器，配置文件格式见./configs/manual/mqttps_manual.yaml",
	"支持自定义上行payload，上行payload文件见./data/payload/mqtt.json",
	"上行数据前需通过页面配置解析脚本, 脚本可在页面上的'产品-数据解析-模板'中找到，或在本仓的./data/mqttpsscript文件中找到",
}

var mqttpsConfS = &conf{}
var mqttpsCmd = &cobra.Command{
	Use:   "mqttps",
	Short: "模拟mqtt透传设备",
	Long:  strings.Join(mqttpsDesc, "\n"),
	Run:   RunMqttpsSim,
}

func init() {
	rootCmd.AddCommand(mqttpsCmd)
	v := newSetting("mqttps_manual")
	v.Unmarshal(&mqttpsConfS)
}

func RunMqttpsSim(cmd *cobra.Command, args []string) {
	customizedPostTopic := fmt.Sprintf("customized/data/pid/%v/devkey/%v", mqttpsConfS.Pid, mqttpsConfS.Did)
	customizedCmdTopic := fmt.Sprintf("customized/cmd/pid/%v/devkey/%v", mqttpsConfS.Pid, mqttpsConfS.Did)
	customizedRespTopic := fmt.Sprintf("customized/resp/pid/%v/devkey/%v", mqttpsConfS.Pid, mqttpsConfS.Did)

	c := mqttClient.EdgeMqttClient{
		ServerHostAndPort: mqttpsConfS.Host,
		ProductID:         strconv.Itoa(mqttpsConfS.Pid),
		DeviceKey:         mqttpsConfS.DevSecret,
		DeviceName:        "some_device",
		DeviceID:          strconv.Itoa(mqttpsConfS.Did),
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
		log.Fatalf("MQTT客户端连接错误：%v", token.Error())
	}

	log.Println("连接成功！等待2秒...")
	time.Sleep(2 * time.Second)

	subTopics := make(map[string]byte)
	subTopics[customizedCmdTopic] = 1

	if token := c.SubscribeMultiple(subTopics, nil); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// 上行数据
	qos := byte(1)
	retained := false
	go func() {
		for {
			pl, err := ioutil.ReadFile(mqttpsConfS.PayloadFile)
			if err != nil {
				log.Fatalf("读取payload文件时发生错误: %v", err)
			}
			if token := c.Publish(customizedPostTopic, qos, retained, pl); token.Wait() && token.Error() != nil {
				log.Fatalf("MQTT客户端Pub错误：%v", token.Error())
			}
			fmt.Printf("已上行数据: %v\n", string(pl))
			// fmt.Printf("已上行数据bytes: %v\n", []byte(pl))
			time.Sleep(time.Duration(mqttpsConfS.Interval) * time.Second)
		}
	}()

	wg.Add(1)
	wg.Wait()

}
