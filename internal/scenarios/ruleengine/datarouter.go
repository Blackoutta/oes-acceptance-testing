package ruleengine

import (
	"oes-acceptance-testing/v1/internal/tester"
	"oes-acceptance-testing/v1/pkg/helpers/conf"
	"oes-acceptance-testing/v1/pkg/helpers/random"
	"strconv"
)

type consumerMessage struct {
	DeviceID string `json:"deviceID,omitempty"`
}

func prepareDatarouter(t *tester.Tester) {
	var err error
	//创建rest地址
	restAddr := tester.CreateDMPRouterAddressPl{
		Name: "rest_addr_" + random.RandString(),
		Type: tester.ROUTER_ADDR_REST,
		Env:  1,
		Host: conf.DataRouterDestination,
		Port: 9005,
		Path: "/rule",
	}
	t.Record, err = t.CreateDMPRouterAddress(restAddr)
	t.Must(err)
	t.AssertTrue("创建rest地址", t.Record.IsSuccess())
	t.Params.RestAddrID = t.Record.Response.(*tester.SingleDataResp).Data

	//创建rest路由
	restRouter := tester.CreateDMPRouterPl{
		Name:            "rest_router_" + random.RandString(),
		Env:             1,
		Type:            tester.ROUTER_TYPE_RULEENGINE,
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

	//启动rest路由
	t.Record, err = t.EnableDMPDataRouter(nil, t.Params.RestRouterID)
	t.Must(err)
	t.AssertTrue("启动rest路由", t.Record.IsSuccess())
}
