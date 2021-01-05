package main

import (
	"crypto/tls"
	"net/http"
	"oes-acceptance-testing/v1/internal/scenarios/datarouter"
	httpscene "oes-acceptance-testing/v1/internal/scenarios/http"
	"oes-acceptance-testing/v1/internal/scenarios/lwm2m"
	"oes-acceptance-testing/v1/internal/scenarios/modbusRTU"
	"oes-acceptance-testing/v1/internal/scenarios/mqtt"
	"oes-acceptance-testing/v1/internal/scenarios/mqttps"
	"oes-acceptance-testing/v1/internal/scenarios/ruleengine"
	"oes-acceptance-testing/v1/internal/scenarios/tcp"
	"oes-acceptance-testing/v1/internal/tester"
	"time"
)

var onHold bool

func init() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func main() {
	resultChan := make(chan tester.TestResult)

	//在这里添加场景
	scenarios := []tester.ScenarioFn{
		lwm2m.RunLwm2mTest,
		datarouter.RunDataRouterTest,
		mqtt.RunMQTTTest,
		mqttps.RunMQTTPSTest,
		tcp.RunTcpTest,
		modbusRTU.RunModbusRTUTest,
		httpscene.RunHTTPTest,
		ruleengine.RunRuleengineTest,
	}

	for _, f := range scenarios {
		go f(resultChan, onHold)
		time.Sleep(5 * time.Second)
	}

	tester.HandleResult(resultChan, len(scenarios))
}
