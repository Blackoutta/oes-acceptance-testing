package tester

// Record records the URL, body and method of a request, as well as the response of that request.
type Record struct {
	Response
	ReqURL         string
	ReqBody        string
	ReqMethod      string
	ResponseString string
}

// Response repersents an http response from the server side
type Response interface {
	IsSuccess() bool
	GetMsg() string
	GetCode() int
}

// GeneralResp is a response struct related to product and devices
type GeneralResp struct {
	Success bool   `json:"success,omitempty"`
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Data    struct {
		ID        int    `json:"id,omitempty"`
		ApiKey    string `json:"apiKey,omitempty"`
		MasterKey string `json:"masterKey,omitempty"`
	} `json:"data,omitempty"`
}

func (r GeneralResp) IsSuccess() bool {
	return r.Success
}

func (r GeneralResp) GetMsg() string {
	return r.Msg
}

func (r GeneralResp) GetCode() int {
	return r.Code
}

// Cmd is a response struct related to issing commands
type CmdResp struct {
	Success bool   `json:"success,omitempty"`
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Data    struct {
		Value interface{} `json:"value,omitempty"`
	} `json:"data,omitempty"`
}

func (r CmdResp) IsSuccess() bool {
	return r.Success
}

func (r CmdResp) GetMsg() string {
	return r.Msg
}

func (r CmdResp) GetCode() int {
	return r.Code
}

// DataPointResp 是跟数据点查询相关的响应体
type DataPointResp struct {
	Success bool   `json:"success,omitempty"`
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Data    []struct {
		Value interface{} `json:"value,omitempty"`
	} `json:"data,omitempty"`
}

func (r DataPointResp) IsSuccess() bool {
	return r.Success
}

func (r DataPointResp) GetMsg() string {
	return r.Msg
}

func (r DataPointResp) GetCode() int {
	return r.Code
}

// DataPointResp 是跟数据点查询相关的响应体
type ContentDataPointResp struct {
	Success bool   `json:"success,omitempty"`
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Data    struct {
		Content []struct {
			Value interface{} `json:"value,omitempty"`
		} `json:"content,omitempty"`
	} `json:"data,omitempty"`
}

func (r ContentDataPointResp) IsSuccess() bool {
	return r.Success
}

func (r ContentDataPointResp) GetMsg() string {
	return r.Msg
}

func (r ContentDataPointResp) GetCode() int {
	return r.Code
}

// SingleDataResp 是跟创建资源相关的响应体，该响应体的data中一般返回创建的资源的ID
type SingleDataResp struct {
	Success bool   `json:"success,omitempty"`
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Data    int    `json:"data,omitempty"`
}

func (r SingleDataResp) IsSuccess() bool {
	return r.Success
}

func (r SingleDataResp) GetMsg() string {
	return r.Msg
}

func (r SingleDataResp) GetCode() int {
	return r.Code
}

// NoDataResp 响应体用来接收不需要检查data的响应体
type NoDataResp struct {
	Success bool   `json:"success,omitempty"`
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
}

func (r NoDataResp) IsSuccess() bool {
	return r.Success
}

func (r NoDataResp) GetMsg() string {
	return r.Msg
}

func (r NoDataResp) GetCode() int {
	return r.Code
}

type IssueAsyncCmdResp struct {
	Success bool   `json:"success,omitempty"`
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Data    struct {
		CommandId string `json:"commandId,omitempty"`
	} `json:"data,omitempty"`
}

func (r IssueAsyncCmdResp) IsSuccess() bool {
	return r.Success
}

func (r IssueAsyncCmdResp) GetMsg() string {
	return r.Msg
}

func (r IssueAsyncCmdResp) GetCode() int {
	return r.Code
}

type GetAsyncCmdResp struct {
	Success bool   `json:"success,omitempty"`
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Data    struct {
		Value  interface{} `json:"value,omitempty"`
		Result int         `json:"result,omitempty"`
	} `json:"data,omitempty"`
}

func (r GetAsyncCmdResp) IsSuccess() bool {
	return r.Success
}

func (r GetAsyncCmdResp) GetMsg() string {
	return r.Msg
}

func (r GetAsyncCmdResp) GetCode() int {
	return r.Code
}

type SingleDataRespStringID struct {
	Success bool   `json:"success,omitempty"`
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Data    string `json:"data,omitempty"`
}

func (r SingleDataRespStringID) IsSuccess() bool {
	return r.Success
}

func (r SingleDataRespStringID) GetMsg() string {
	return r.Msg
}

func (r SingleDataRespStringID) GetCode() int {
	return r.Code
}

// GeneralResp is a response struct related to product and devices
type MailResp struct {
	Success bool   `json:"success,omitempty"`
	Code    int    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Data    struct {
		ID string `json:"id,omitempty"`
	} `json:"data,omitempty"`
}

func (r MailResp) IsSuccess() bool {
	return r.Success
}

func (r MailResp) GetMsg() string {
	return r.Msg
}

func (r MailResp) GetCode() int {
	return r.Code
}
