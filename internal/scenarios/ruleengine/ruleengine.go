package ruleengine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"oes-acceptance-testing/v1/internal/tester"
	"oes-acceptance-testing/v1/pkg/helpers/conf"
	"oes-acceptance-testing/v1/pkg/helpers/logger"
	"oes-acceptance-testing/v1/pkg/helpers/random"
	"strconv"
	"sync"
	"time"
)

var wg sync.WaitGroup

const (
	suiteName = "场景联动测试套件"
	td        = true
)

func RunRuleengineTest(resultChan chan tester.TestResult, onHold bool) {
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

	// recover if panic, then proceed to teardown
	defer t.Recover()

	// 准备工作
	triggerChan := make(chan int, 1)

	// 启动MQTT设备A
	c1, pid1, did1 := bootMqttClient(t, triggerChan)
	postTopic := fmt.Sprintf("$sys/%v/%v/thing/property/post", t.Params.ProductID, t.Params.DeviceID)

	// 启动MQTT设备B
	_, pid2, did2 := bootMqttClient(t, triggerChan)

	// 创建规则转发用路由
	prepareDatarouter(t)

	// 创建规则1 - 设备A上行触发命令下发给设备B
	rule1 := tester.CreateRulePl{
		Name: "upward_rule_" + random.RandString(),
		Conditions: []tester.RuleCondition{
			{
				ProductId: pid1,
				Parameter: "someDouble",
				Operation: ">",
				Operand:   1,
			},
		},
		TimeTrigger: "",
		Filters:     nil,
		Actions: []tester.RuleAction{
			{
				ProductId:  pid2,
				DeviceId:   did2,
				Identifier: "someBoolean",
				Value:      true,
			},
		},
		Subscription:   nil,
		ExportClientId: 0,
		MailServerId:   "",
		MailPatternId:  "",
		RuleType:       1,
	}

	t.Record, err = t.CreateRule(rule1)
	t.Must(err)
	t.AssertTrue("创建规则1 - 设备A上行触发命令下发给设备B", t.Record.IsSuccess())
	ruleID1 := t.Record.Response.(*tester.SingleDataRespStringID).Data

	// 启动规则1 - 设备A上行触发命令下发给设备B
	t.Record, err = t.EnableRule(ruleID1)
	t.Must(err)
	t.AssertTrue("启动规则1 - 设备A上行触发命令下发给设备B", t.Record.IsSuccess())

	// 创建规则2 - 设备A上行触发转发到Rest服务器
	rule2 := tester.CreateRulePl{
		Name: "forward_rule_" + random.RandString(),
		Conditions: []tester.RuleCondition{
			{
				ProductId: pid1,
				Parameter: "someDouble",
				Operation: ">",
				Operand:   1,
			},
		},
		TimeTrigger:    "",
		Filters:        nil,
		Subscription:   nil,
		ExportClientId: t.Params.RestRouterID,
		MailServerId:   "",
		MailPatternId:  "",
		RuleType:       1,
	}

	t.Record, err = t.CreateRule(rule2)
	t.Must(err)
	t.AssertTrue("创建规则2 - 设备A上行触发转发到Rest服务器", t.Record.IsSuccess())
	ruleID2 := t.Record.Response.(*tester.SingleDataRespStringID).Data

	// 启动规则2 - 设备A上行触发转发到Rest服务器
	t.Record, err = t.EnableRule(ruleID2)
	t.Must(err)
	t.AssertTrue("启动规则2 - 设备A上行触发转发到Rest服务器", t.Record.IsSuccess())

	// 创建邮件服务器
	mailServer := tester.CreateMailServerPl{
		Name:       "mail_server_" + random.RandString(),
		Account:    "179278445@qq.com",
		AuthCode:   "sbgnkbfdmqrjbhjc",
		ServerHost: "smtp.qq.com",
		ServerPort: "465",
		Type:       1,
	}
	t.Record, err = t.CreateMailServer(mailServer)
	t.Must(err)
	t.AssertTrue("创建邮件服务器", t.Record.IsSuccess())
	mailServerID := t.Record.Response.(*tester.MailResp).Data.ID

	// 创建邮件模板
	bs, err := ioutil.ReadFile("data/mailtemplate")
	if err != nil {
		err := fmt.Sprintf("error while reading mail template file: %v", err)
		panic(err)
	}
	mailTemplate := tester.CreateMailTemplate{
		Name:    "mail_template_" + random.RandString(),
		Content: string(bs),
		Type:    1,
	}
	t.Record, err = t.CreateMailTemplate(mailTemplate)
	t.Must(err)
	t.AssertTrue("创建邮件模板", t.Record.IsSuccess())
	mailTemplateID := t.Record.Response.(*tester.MailResp).Data.ID

	// 创建规则4 - 设备上行触发发送邮件
	rule4 := tester.CreateRulePl{
		Name: "qq_mail_rule_" + random.RandString(),
		Conditions: []tester.RuleCondition{
			{
				ProductId: pid1,
				Parameter: "someDouble",
				Operation: ">",
				Operand:   1,
			},
		},
		TimeTrigger:   "",
		Filters:       nil,
		Subscription:  []string{"179278445@qq.com"},
		MailServerId:  mailServerID,
		MailPatternId: mailTemplateID,
		RuleType:      1,
	}

	t.Record, err = t.CreateRule(rule4)
	t.Must(err)
	fmt.Printf("新建的邮件规则Body：\n%v\n", t.Record.ReqBody)
	fmt.Printf("新建的邮件规则响应：\n%v\n", t.Record.ResponseString)
	t.AssertTrue("创建规则4 - 设备A上行触发发送邮件", t.Record.IsSuccess())
	ruleID4 := t.Record.Response.(*tester.SingleDataRespStringID).Data

	//启动规则4 - 设备A上行触发发送邮件
	t.Record, err = t.EnableRule(ruleID4)
	t.Must(err)
	t.AssertTrue("启动规则4 - 设备A上行触发发送邮件", t.Record.IsSuccess())

	// 创建规则5 - 设备上行触发发送邮件(从公司邮箱发送)
	rule5 := tester.CreateRulePl{
		Name: "company_mail_rule_" + random.RandString(),
		Conditions: []tester.RuleCondition{
			{
				ProductId: pid1,
				Parameter: "someDouble",
				Operation: ">",
				Operand:   1,
			},
		},
		TimeTrigger:  "",
		Filters:      nil,
		Subscription: []string{"179278445@qq.com"},
		RuleType:     1,
	}

	t.Record, err = t.CreateRule(rule5)
	t.Must(err)
	t.AssertTrue("创建规则5 - 设备上行触发发送邮件(从公司邮箱发送)", t.Record.IsSuccess())
	ruleID5 := t.Record.Response.(*tester.SingleDataRespStringID).Data

	//启动规则5 - 设备上行触发发送邮件(从公司邮箱发送)
	t.Record, err = t.EnableRule(ruleID5)
	t.Must(err)
	t.AssertTrue("启动规则5 - 设备上行触发发送邮件(从公司邮箱发送)", t.Record.IsSuccess())

	time.Sleep(2 * time.Second)

	// 设备A上行数据
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

	fmt.Println("上行数据中...")

	go func() {
		for {
			if token := c1.Publish(postTopic, 1, false, pl); token.Wait() && token.Error() != nil {
				t.Logger.Fatalf("MQTT客户端Pub错误：%v", token.Error())
			}
			fmt.Println("已上行数据...")
			time.Sleep(2 * time.Second)
		}
	}()

	// 检查规则1触发
	t.Logger.Println("检查规则1触发中...10秒后超时")
	select {
	case <-triggerChan:
		t.Logger.Println("收到设备触发的命令.")
		t.Logger.Println("规则1 - 设备A上行触发命令下发给设备B ......Pass!")
	case <-time.After(10 * time.Second):
		t.FailTest("超过10秒未收到设备触发命令命令，测试失败！")
	}

	t.Logger.Println("检查规则2触发中...")
	time.Sleep(3 * time.Second)
	// 检查规则2触发
	var restMsg consumerMessage
	resp, err := http.Get(fmt.Sprintf("http://%v:9005/getdataForRule", conf.DataRouterDestination))
	if err != nil {
		panic(err)
	}
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bs, &restMsg); err != nil {
		restErr := fmt.Errorf("error while unmarshaling rest msg: %v", err)
		panic(restErr)
	}
	if restMsg.DeviceID == strconv.Itoa(did1) {
		t.Logger.Println("规则2 - 设备A上行触发转发到Rest服务器 ......Pass!")
	} else {
		t.FailTest("设备A上行触发转发到Rest服务器 ......Fail!")
	}
	// 创建规则 - 定时触发命令下发
	rule3 := tester.CreateRulePl{
		Name:          "timed_rule_" + random.RandString(),
		TimeTrigger:   "*/2 * * * * ?",
		Filters:       nil,
		Subscription:  nil,
		MailServerId:  "",
		MailPatternId: "",
		RuleType:      1,
		Actions: []tester.RuleAction{
			{
				ProductId:  pid2,
				DeviceId:   did2,
				Identifier: "someBoolean",
				Value:      true,
			},
		},
	}

	t.Record, err = t.CreateRule(rule3)
	t.Must(err)
	t.AssertTrue("创建规则3 - 定时触发命令下发", t.Record.IsSuccess())
	ruleID3 := t.Record.Response.(*tester.SingleDataRespStringID).Data

	// 启动规则3 - 定时触发命令下发
	t.Record, err = t.EnableRule(ruleID3)
	t.Must(err)
	t.AssertTrue("启动规则3 - 定时触发命令下发", t.Record.IsSuccess())

	t.Logger.Println("检查规则3触发中...10秒后超时")
	// 检查规则3触发
	select {
	case <-triggerChan:
		t.Logger.Println("收到定时触发命令.")
		t.Logger.Println("规则3 - 定时触发命令下发 ......Pass!")
	case <-time.After(15 * time.Second):
		t.FailTest("超过10秒未触发定时命令，测试失败！")
	}

	// 创建规则 - 设备上线发送命令
	// 创建规则 - 设备下线发送邮件
	// 创建规则 - 设备删除发送邮件

	if onHold == true {
		wg.Add(1)
		wg.Wait()
	}

	//删除产品1
	t.Record, err = t.DeleteDmpProduct(pid1)
	t.Must(err)
	t.AssertTrue("删除MQTT产品1", t.Record.IsSuccess())

	//删除产品2
	t.Record, err = t.DeleteDmpProduct(pid2)
	t.Must(err)
	t.AssertTrue("删除MQTT产品2", t.Record.IsSuccess())

	//删除规则1
	t.Record, err = t.DeleteRule(ruleID1)
	t.Must(err)
	t.AssertTrue("删除规则1", t.Record.IsSuccess())

	//删除规则2
	t.Record, err = t.DeleteRule(ruleID2)
	t.Must(err)
	t.AssertTrue("删除规则2", t.Record.IsSuccess())

	//删除规则3
	t.Record, err = t.DeleteRule(ruleID3)
	t.Must(err)
	t.AssertTrue("删除规则3", t.Record.IsSuccess())

	//禁用rest路由
	t.Record, err = t.DisableDMPDataRouter(nil, t.Params.RestRouterID)
	t.Must(err)
	t.AssertTrue("禁用rest路由", t.Record.IsSuccess())

	//删除路由
	t.Record, err = t.DeleteDMPDataRouter(nil, t.Params.RestRouterID)
	t.Must(err)
	t.AssertTrue("删除rest路由", t.Record.IsSuccess())

	//删除地址
	t.Record, err = t.DeleteDMPRouterAddress(nil, t.Params.RestAddrID)
	t.Must(err)
	t.AssertTrue("删除rest地址", t.Record.IsSuccess())

	//判断套件是否成功并打印结果
	t.DisplayResult(resultChan)
}
