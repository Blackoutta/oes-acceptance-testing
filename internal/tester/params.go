package tester

type Command struct {
	FunctionType string `json:"functionType"`
	UUID         string `json:"uuid"`
	Identifier   string `json:"identifier"`
}

type GetCommand struct {
	Id           string   `json:"id"`
	Version      string   `json:"version"`
	ProductId    int      `json:"productId"`
	DeviceId     int      `json:"deviceId"`
	FunctionType string   `json:"functionType"`
	Params       []string `json:"params"`
	Timeout      int      `json:"timeout"`
	Mode         int      `json:"mode"`
}

type SetCommand struct {
	Id           string                 `json:"id"`
	Version      string                 `json:"version"`
	ProductId    int                    `json:"productId"`
	DeviceId     int                    `json:"deviceId"`
	FunctionType string                 `json:"functionType"`
	Params       map[string]interface{} `json:"params"`
	Timeout      int                    `json:"timeout"`
	Mode         int                    `json:"mode"`
}

type Params struct {
	ProductID int
	DeviceID  int
	ApiKey    string
	MasterKey string
	EdgeCmd   Command

	RestAddrID  int
	MQTTAddrID  int
	KAFKAAddrID int
	MYSQLAddrID int

	RestRouterID  int
	KafkaRouterID int
	MQTTRouterID  int
	MYSQLRouterID int

	ModbusProdID    int
	ModbusDeviceID  int
	ModbusMasterKey string
	ModbusApiKey    string
}
