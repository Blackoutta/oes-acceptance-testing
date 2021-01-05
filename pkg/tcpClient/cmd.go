package tcpClient

type cmdResp struct {
	Uuid            string      `json:"uuid"`
	IdentifierValue interface{} `json:"identifierValue"`
}

type tcpCommand struct {
	Uuid         string
	FunctionType string
	Params       map[string]interface{}
}
