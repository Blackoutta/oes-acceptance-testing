package tester

import (
	"fmt"
	"net/http"
	"strconv"
)

func (t *Tester) CreateDmpProduct(pl interface{}) (Record, error) {
	path := fmt.Sprintf("/products")
	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, nil)
	return t.Send(&GeneralResp{})
}

func (t *Tester) DeleteDmpProduct(productID int) (Record, error) {
	path := fmt.Sprintf("/products/%v", productID)
	t.Request = NewRequest(path, http.MethodDelete, PLATFORM_DMP, nil, nil)
	return t.Send(&GeneralResp{})
}

func (t *Tester) CreateDmpProperty(productID int, pl interface{}) (Record, error) {
	path := fmt.Sprintf("/products/%v/properties", productID)
	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, nil)
	return t.Send(&GeneralResp{})
}

func (t *Tester) CreateDmpDevice(productId int, pl interface{}) (Record, error) {
	path := fmt.Sprintf("/devices")

	q := URLQuery{}
	q.Add("productId", strconv.Itoa(productId))

	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, q)
	return t.Send(&GeneralResp{})
}

func (t *Tester) CreateLWM2MDevice(productId int, pl interface{}) (Record, error) {
	path := fmt.Sprintf("/devices/lwm2m")

	q := URLQuery{}
	q.Add("productId", strconv.Itoa(productId))

	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, q)
	return t.Send(&GeneralResp{})
}

func (t *Tester) CreateTcpScript(productId int, masterKey string, pl interface{}) (Record, error) {
	path := fmt.Sprintf("/scripts")

	q := URLQuery{}
	q["productId"] = strconv.Itoa(productId)
	q["masterKey"] = masterKey

	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, q)
	return t.Send(&GeneralResp{})
}

func (t *Tester) IssueTcpCommand(pl interface{}) (Record, error) {
	path := fmt.Sprintf("/commands/tcp")

	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, nil)
	return t.Send(&CmdResp{})
}

func (t *Tester) IssueMQTTCommand(pl interface{}) (Record, error) {
	path := fmt.Sprintf("/commands/mqtt")

	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, nil)
	return t.Send(&CmdResp{})
}

func (t *Tester) GetLatestDataPoints(pl interface{}, deviceID, limit int, resourceName string) (Record, error) {
	path := fmt.Sprintf("/data/device/%v/limit/%v", deviceID, limit)
	q := URLQuery{}
	q.Add("resourceName", resourceName)
	t.Request = NewRequest(path, http.MethodGet, PLATFORM_DMP, pl, q)
	return t.Send(&DataPointResp{})
}

func (t *Tester) CreateDMPRouterAddress(pl interface{}) (Record, error) {
	path := fmt.Sprintf("/address-configs")
	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, nil)
	return t.Send(&SingleDataResp{})
}

func (t *Tester) CreateDMPDataRouter(pl interface{}) (Record, error) {
	path := fmt.Sprintf("/routers")
	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, nil)
	return t.Send(&SingleDataResp{})
}

func (t *Tester) CreateModbusProperty(pl interface{}, productID int) (Record, error) {
	path := fmt.Sprintf("/products/%v/properties/modbus", productID)
	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, nil)
	return t.Send(&GeneralResp{})
}

func (t *Tester) AssociateDTUWithModbusDevice(pl interface{}, dtuId int, dtuKey string, productId int, masterKey string, deviceIds string) (Record, error) {
	path := fmt.Sprintf("/devices/dtu/associations/modbus")
	q := URLQuery{}
	q.Add("dtuId", strconv.Itoa(dtuId))
	q.Add("dtuKey", dtuKey)
	q.Add("productId", strconv.Itoa(productId))
	q.Add("masterKey", masterKey)
	q.Add("deviceIds", deviceIds)

	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, q)
	return t.Send(&NoDataResp{})
}

func (t *Tester) EnableDMPDataRouter(pl interface{}, routerID int) (Record, error) {
	path := fmt.Sprintf("/routers/%v/enable", routerID)
	t.Request = NewRequest(path, http.MethodPut, PLATFORM_DMP, pl, nil)
	return t.Send(&NoDataResp{})
}

func (t *Tester) DisableDMPDataRouter(pl interface{}, routerID int) (Record, error) {
	path := fmt.Sprintf("/routers/%v/disable", routerID)
	t.Request = NewRequest(path, http.MethodPut, PLATFORM_DMP, pl, nil)
	return t.Send(&NoDataResp{})
}

func (t *Tester) DeleteDMPDataRouter(pl interface{}, routerID int) (Record, error) {
	path := fmt.Sprintf("/routers/%v", routerID)
	t.Request = NewRequest(path, http.MethodDelete, PLATFORM_DMP, pl, nil)
	return t.Send(&NoDataResp{})
}

func (t *Tester) DeleteDMPRouterAddress(pl interface{}, addressID int) (Record, error) {
	path := fmt.Sprintf("/address-configs/%v", addressID)
	t.Request = NewRequest(path, http.MethodDelete, PLATFORM_DMP, pl, nil)
	return t.Send(&NoDataResp{})
}

func (t *Tester) IssueLWM2MCommand(pl interface{}) (Record, error) {
	path := fmt.Sprintf("/commands/lwm2m")
	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, nil)
	return t.Send(&CmdResp{})
}

func (t *Tester) AddModbusChannelToDTU(pl interface{}) (Record, error) {
	path := fmt.Sprintf("/devices/modbus/channel-config/dtu")
	t.Request = NewRequest(path, http.MethodPut, PLATFORM_DMP, pl, nil)
	return t.Send(&NoDataResp{})
}

func (t *Tester) IssueModbusDTUCommandByName(pl interface{}) (Record, error) {
	path := fmt.Sprintf("/device/command/byname/dtu/async")
	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, nil)
	return t.Send(&IssueAsyncCmdResp{})
}

func (t *Tester) GetAsyncCommandResult(pl interface{}, cmdID string) (Record, error) {
	path := fmt.Sprintf("/commands-async/%s", cmdID)
	t.Request = NewRequest(path, http.MethodGet, PLATFORM_DMP, pl, nil)
	return t.Send(&GetAsyncCmdResp{})
}

func (t *Tester) GetMultipleLatestDataPoints(deviceId int, propertyNames []string) (Record, error) {
	path := fmt.Sprintf("/data/properties/recent/value")
	q := URLQuery{}
	q.Add("deviceId", strconv.Itoa(deviceId))
	var names string
	for i, v := range propertyNames {
		if i < len(propertyNames)-1 {
			names += v + ","
		} else {
			names += v
		}
	}

	q.Add("propertyNames", names)
	t.Request = NewRequest(path, http.MethodGet, PLATFORM_DMP, nil, q)
	return t.Send(&DataPointResp{})
}

func (t *Tester) GetDataPointsInTimePeriod(deviceId int, resourceName string, start, end int64, currentPage, pageSize int) (Record, error) {
	path := fmt.Sprintf("/data/device/%v", deviceId)
	q := URLQuery{}
	q.Add("deviceId", strconv.Itoa(deviceId))
	q.Add("resourceName", resourceName)
	q.Add("start", strconv.Itoa(int(start)))
	q.Add("end", strconv.Itoa(int(end)))
	q.Add("currentPage", strconv.Itoa(currentPage))
	q.Add("pageSize", strconv.Itoa(pageSize))
	t.Request = NewRequest(path, http.MethodGet, PLATFORM_DMP, nil, q)
	return t.Send(&ContentDataPointResp{})
}

func (t *Tester) CreateRule(pl interface{}) (Record, error) {
	path := fmt.Sprintf("/rules")
	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, nil)
	return t.Send(&SingleDataRespStringID{})
}

func (t *Tester) EnableRule(ruleId string) (Record, error) {
	path := fmt.Sprintf("/rules/%v/enable", ruleId)
	t.Request = NewRequest(path, http.MethodPut, PLATFORM_DMP, nil, nil)
	return t.Send(&SingleDataRespStringID{})
}

func (t *Tester) DeleteRule(ruleId string) (Record, error) {
	path := fmt.Sprintf("/rules/%v", ruleId)
	t.Request = NewRequest(path, http.MethodDelete, PLATFORM_DMP, nil, nil)
	return t.Send(&NoDataResp{})
}

func (t *Tester) CreateMailServer(pl interface{}) (Record, error) {
	path := fmt.Sprintf("/mail/server/add")
	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, nil)
	return t.Send(&MailResp{})
}

func (t *Tester) CreateMailTemplate(pl interface{}) (Record, error) {
	path := fmt.Sprintf("/mail/pattern/add")
	t.Request = NewRequest(path, http.MethodPost, PLATFORM_DMP, pl, nil)
	return t.Send(&MailResp{})
}
