package cmd

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"oes-acceptance-testing/v1/pkg/mqttClient"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var mqttDesc = []string{
	"mqtt设备模拟器，配置文件格式见./configs/manual/mqtt_manual.yaml",
	"支持自定义上行payload，上行payload文件见./data/payload/mqtt.json",
}

var dynamicRegister bool
var configFile string

var mqttConfS = &conf{}
var mqttCmd = &cobra.Command{
	Use:   "mqtt",
	Short: "模拟mqtt设备",
	Long:  strings.Join(mqttDesc, "\n"),
	Run:   RunMqttSim,
}

func init() {
	rootCmd.AddCommand(mqttCmd)
	mqttCmd.Flags().BoolVarP(&dynamicRegister, "dynamic-register", "d", false, "使用动态注册, 如使用该选项，需首先需在页面上产品详情页中开启动态注册）")
	mqttCmd.Flags().StringVarP(&configFile, "config", "c", "mqtt_manual", "指定配置文件名（无需后缀扩展名），默认为mqtt_manual")
}

func RunMqttSim(cmd *cobra.Command, args []string) {
	v := newSetting(configFile)
	v.Unmarshal(&mqttConfS)
	log.Printf("使用了配置文件: %v\n", configFile)

	setTopic := fmt.Sprintf("$sys/%v/%v/thing/property/set", mqttConfS.Pid, mqttConfS.Did)
	setReplyTopic := fmt.Sprintf("$sys/%v/%v/thing/property/set_reply", mqttConfS.Pid, mqttConfS.Did)
	postReplyTopic := fmt.Sprintf("$sys/%v/%v/thing/property/post/reply", mqttConfS.Pid, mqttConfS.Did)
	postTopic := fmt.Sprintf("$sys/%v/%v/thing/property/post", mqttConfS.Pid, mqttConfS.Did)
	shadowUpdateRespTopic := fmt.Sprintf("shadow/update/resp/pid/%v/devkey/%v", mqttConfS.Pid, mqttConfS.Did)
	shadowGetResp := fmt.Sprintf("shadow/get/resp/pid/%v/devkey/%v", mqttConfS.Pid, mqttConfS.Did)
	configTopic := fmt.Sprintf("$sys/%v/%v/thing/config/push", mqttConfS.Pid, mqttConfS.Did)
	ntpRespTopic := fmt.Sprintf("/ext/ntp/%v/%v/response", mqttConfS.Pid, mqttConfS.Did)

	c := &mqttClient.EdgeMqttClient{
		ServerHostAndPort: mqttConfS.Host,
		ProductID:         strconv.Itoa(mqttConfS.Pid),
		DeviceKey:         mqttConfS.DevSecret,
		DeviceName:        "mqtt_device",
		DeviceID:          strconv.Itoa(mqttConfS.Did),
		AuthMsg: mqttClient.AuthMessage{
			SasToken: mqttClient.SasToken{
				Version: "2018-10-31",
				Et:      time.Now().AddDate(0, 3, 0).Unix(),
				Method:  mqttClient.AUTH_SHA1,
			},
		},
	}

	if dynamicRegister {
		log.Println("使用了动态注册...")
		// 进行动态注册
		dynamicRegisterURL := fmt.Sprintf("http://%v/dynamicregister", mqttConfS.DynamicRegisterServer)
		fmt.Println("注册URL:", dynamicRegisterURL)
		registerClient := &http.Client{
			Timeout: 10 * time.Second,
		}

		drToken, err := DynamicRegisterToken(mqttConfS.Pid, mqttConfS.ProdSecret, mqttConfS.DevName)
		// drToken, err := DynamicRegisterToken(101181, "ZDNjZjkwMTgyZDFkMDc2Yzg2NTE=", "mqtt_device")
		if err != nil {
			fmt.Printf("生成动态注册token时发生错误: %v\n", err)
		}
		fmt.Println("dr token: " + drToken)

		dr := fmt.Sprintf(`{"deviceName": "%v", "pid": "%v", "token": "%v"}`, mqttConfS.DevName, mqttConfS.Pid, drToken)
		fmt.Println("注册body:" + dr)

		registerResp, err := registerClient.Post(dynamicRegisterURL, "application/json", bytes.NewReader([]byte(dr)))
		if err != nil {
			fmt.Printf("请求产品级鉴权时发生错误: %v\n", err)
		}

		registerRespBody := make([]byte, 0)
		if registerResp != nil {
			registerRespBody, err = ioutil.ReadAll(registerResp.Body)
			if err != nil {
				fmt.Printf("读取动态注册响应时发生错误: %v\n", err)
			}

			fmt.Printf("动态注册获取到设备信息: %v\n", string(registerRespBody))
			registerResp.Body.Close()
		}
		drResp := struct {
			Code int `json:"code"`
			Data struct {
				DeviceId     int    `json:"deviceId"`
				DeviceSecret string `json:"deviceSecret"`
			} `json:"data"`
		}{}

		if err := json.Unmarshal(registerRespBody, &drResp); err != nil {
			fmt.Printf("解析动态注册响应失败: %v\n", err)
			os.Exit(-1)
		}

		if drResp.Code != 0 {
			fmt.Println("动态注册失败，程序结束...")
			os.Exit(-1)
		}

		c.DeviceID = fmt.Sprintf("%d", drResp.Data.DeviceId)
		c.DeviceKey = drResp.Data.DeviceSecret
	}

	if err := c.NewMqttClient(600, mqttClient.NewMqttHandler(c.ProductID, c.DeviceID, setReplyTopic)); err != nil {
		panic(err)
	}

	// 设备鉴权+上线
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT客户端连接错误：%v", token.Error())
	}

	log.Println("连接成功！等待2秒...")
	time.Sleep(2 * time.Second)

	subTopics := make(map[string]byte)
	subTopics[setTopic] = 1
	subTopics[postReplyTopic] = 1
	subTopics[configTopic] = 1
	subTopics[ntpRespTopic] = 1
	subTopics[shadowGetResp] = 1
	subTopics[shadowUpdateRespTopic] = 1

	if token := c.SubscribeMultiple(subTopics, nil); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// 上行数据
	qos := mqttConfS.Qos
	retained := false
	go func() {
		counter := 1
		for {
			pl, err := ioutil.ReadFile(mqttConfS.PayloadFile)
			if err != nil {
				log.Fatalf("读取payload文件时发生错误: %v", err)
			}
			if token := c.Publish(postTopic, qos, retained, pl); token.Wait() && token.Error() != nil {
				log.Fatalf("MQTT客户端Pub错误：%v", token.Error())
			}
			fmt.Printf("已上行数据: %v\n", string(pl))
			fmt.Println("已经发送条数：", counter)
			counter++
			// fmt.Printf("已上行数据bytes: %v\n", []byte(pl))
			time.Sleep(time.Duration(mqttConfS.Interval) * time.Second)
		}
	}()

	// 根据键盘输入执行各种能力
	go func() {
		var cmd string
		for {
			_, err := fmt.Scanln(&cmd)
			if err != nil {
				if err.Error() == "unexpected newline" {
					continue
				}
				fmt.Println(err)
			}

			switch cmd {
			case "u":
				postShadowTopic := fmt.Sprintf("shadow/update/pid/%v/devkey/%v", c.ProductID, c.DeviceID)
				postShadowPl, err := ioutil.ReadFile(mqttConfS.ShadowFile)
				if err != nil {
					panic(err)
				}
				log.Printf("上报shadow至topic: %v\n", postShadowTopic)
				log.Printf("上报shadow payload:\n%v\n", string(postShadowPl))
				if token := c.Publish(postShadowTopic, qos, retained, postShadowPl); token.Wait() && token.Error() != nil {
					log.Printf("MQTT客户端Pub错误：%v", token.Error())
					continue
				}
			case "g":
				getShadowPl := `{"method": "get"}`
				getShadowTopic := fmt.Sprintf("shadow/get/pid/%v/devkey/%v", c.ProductID, c.DeviceID)
				log.Printf("发送获取shadow请求至至topic: %v\n", getShadowTopic)
				if token := c.Publish(getShadowTopic, qos, retained, getShadowPl); token.Wait() && token.Error() != nil {
					log.Printf("MQTT客户端Pub错误：%v", token.Error())
					continue
				}
			case "n":
				ntpPl := fmt.Sprintf(`{"deviceSendTime": %v}`, time.Now().Unix()*1000)
				ntpTopic := fmt.Sprintf(`/ext/ntp/%v/%v/request`, c.ProductID, c.DeviceID)
				fmt.Printf("发送时钟同步请求至topic: %v\n", ntpTopic)
				if token := c.Publish(ntpTopic, qos, retained, ntpPl); token.Wait() && token.Error() != nil {
					log.Printf("MQTT客户端Pub错误：%v", token.Error())
					continue
				}
			}
		}
	}()

	wg.Add(1)
	wg.Wait()

}

func DynamicRegisterToken(pid int, pkey string, deviceName string) (string, error) {
	target := strconv.Itoa(pid) + "&" + deviceName
	mac := hmac.New(sha1.New, []byte(pkey))
	_, err := mac.Write([]byte(target))
	if err != nil {
		return "", err
	}
	src := mac.Sum(nil)
	dst := url.QueryEscape(base64.StdEncoding.EncodeToString(src))
	return dst, nil
}
