package datarouter

import "oes-acceptance-testing/v1/internal/tester"

func tearDown(t *tester.Tester, resultChan chan tester.TestResult) {
	//删除产品
	var err error
	t.Record, err = t.DeleteDmpProduct(t.Params.ProductID)
	t.Must(err)
	t.AssertTrue("删除MQTT产品", t.Record.IsSuccess())

	//删除路由
	t.Record, err = t.DeleteDMPDataRouter(nil, t.Params.RestRouterID)
	t.Must(err)
	t.AssertTrue("删除rest路由", t.Record.IsSuccess())

	t.Record, err = t.DeleteDMPDataRouter(nil, t.Params.MQTTRouterID)
	t.Must(err)
	t.AssertTrue("删除mqtt路由", t.Record.IsSuccess())

	t.Record, err = t.DeleteDMPDataRouter(nil, t.Params.MYSQLRouterID)
	t.Must(err)
	t.AssertTrue("删除mysql路由", t.Record.IsSuccess())

	t.Record, err = t.DeleteDMPDataRouter(nil, t.Params.KafkaRouterID)
	t.Must(err)
	t.AssertTrue("删除kafka路由", t.Record.IsSuccess())

	//删除地址
	t.Record, err = t.DeleteDMPRouterAddress(nil, t.Params.RestAddrID)
	t.Must(err)
	t.AssertTrue("删除rest地址", t.Record.IsSuccess())

	t.Record, err = t.DeleteDMPRouterAddress(nil, t.Params.MQTTAddrID)
	t.Must(err)
	t.AssertTrue("删除mqtt地址", t.Record.IsSuccess())

	t.Record, err = t.DeleteDMPRouterAddress(nil, t.Params.KAFKAAddrID)
	t.Must(err)
	t.AssertTrue("删除kafka地址", t.Record.IsSuccess())

	t.Record, err = t.DeleteDMPRouterAddress(nil, t.Params.MYSQLAddrID)
	t.Must(err)
	t.AssertTrue("删除mysql地址", t.Record.IsSuccess())
	//判断套件是否成功并打印结果
	t.DisplayResult(resultChan)
}
