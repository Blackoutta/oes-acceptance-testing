package main

import (
	"crypto/tls"
	"net/http"
	httpscene "oes-acceptance-testing/v1/internal/scenarios/http"
	"oes-acceptance-testing/v1/internal/tester"
)

var onHold bool

func init() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func main() {
	tester.SetOnhold(&onHold)

	resultChan := make(chan tester.TestResult)

	//在这里添加场景
	scenarios := []tester.ScenarioFn{
		httpscene.RunHTTPTest,
	}

	for _, f := range scenarios {
		go f(resultChan, onHold)
	}

	tester.HandleResult(resultChan, len(scenarios))
}
