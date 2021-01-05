package tester

const (
	PROTOCOL_MQTT         = 1
	PROTOCOL_LWM2M        = 2
	PROTOCOL_TCP          = 3
	PROTOCOL_MODBUS       = 4
	PROTOCOL_HTTP         = 6
	NODETYPE_DEVICE       = 1
	NODETYPE_DTU          = 3
	MODEL_OBJECTMODEL     = 1
	MODEL_NONE            = 0
	DATAFORMAT_STANDARD   = 1
	DATAFORMAT_CUSTOMIZED = 2
	DATAFORMAT_MODBUSRTU  = 3
	AUTHMETHOD_DEVICEKEY  = 1
	AUTHMETHOD_CERT       = 2
	ACCESSMODE_RONLY      = 1
	ACCESSMODE_RW         = 2
)

type CreateDmpProductPl struct {
	Name                 string `json:"name,omitempty"`
	Description          string `json:"description,omitempty"`
	ProtocolType         int    `json:"protocolType,omitempty"`         // 	产品协议 （1 : MQTT; 2 : LWM2M; 3 : TCP; 4 : MODBUS; 6 : HTTP）
	NodeType             int    `json:"nodeType,omitempty"`             //节点类型 (1 : 设备; 3 : DTU)，默认设备类型。
	Model                int    `json:"model,omitempty"`                // 	产品模式 （1：物模型； 0：无） // protocolType为LWM2M时不需要此字段
	DataFormat           int    `json:"dataFormat,omitempty"`           //数据类型 （1：标准数据类型 ； 2：自定义数据类型； 3 : modbus rtu） //protocolType为LWM2M/Http时不需要此字段。
	AuthenticationMethod int    `json:"authenticationMethod,omitempty"` // 	产品认证方式 （1：设备密钥； 2：证书认证）
	DynamicRegister      int    `json:"dynamicRegister"`
	NetworkMethod        int    `json:"networkMethod"`
}

const (
	DATATYPE_STRING = 1 + iota
	DATATYPE_BOOLEAN
	DATATYPE_INT32
	DATATYPE_INT64
	DATATYPE_FLOAT32
	DATATYPE_FLOAT64
	DATATYPE_BYTES
	DATATYPE_DATE
	DATATYPE_ENUM
)

type CreateDmpPropertyPl struct {
	Name       string      `json:"name,omitempty"`
	Identifier string      `json:"identifier,omitempty"`
	AccessMode int         `json:"accessMode,omitempty"`
	Type       int         `json:"type,omitempty"`
	Unit       string      `json:"unit,omitempty"`
	Minimum    interface{} `json:"minimum,omitempty"`
	Maximum    interface{} `json:"maximum,omitempty"`
	Special    Special     `json:"special,omitempty"`
}

type Special struct {
	Length     int         `json:"length,omitempty"` //string类型的字符串长度，1~2048
	Step       float64     `json:"step,omitempty"`
	EnumArrays []EnumArray `json:"enumArray,omitempty"` //枚举或布尔类型的相关信息的数组（布尔类型时元素的key有且仅有0/1，枚举类型时元素的个数不超过20个，key范围0~127）
}

type EnumArray struct {
	Key      int    `json:"key"`                //枚举或布尔类型的值（布尔类型时key有且仅有0/1，枚举类型时key范围0~127）
	Describe string `json:"describe,omitempty"` //值对应的说明
}

type CreateDmpDevicePl struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

const (
	LWM2M_SUBSCRIPTION_ENABLED  = 1
	LWM2M_SUBSCRIPTION_DISABLED = 0
)

type CreateLWM2MDevicePl struct {
	Name         string `json:"name,omitempty"`
	IMEI         string `json:"imei,omitempty"`
	IMSI         string `json:"imsi,omitempty"`
	AuthKey      string `json:"authKey,omitempty"`
	Subscription int    `json:"subscription,omitempty"`
	PSK          string `json:"psk,omitempty"`
	Description  string `json:"description,omitempty"`
}

type CreateTcpScriptPl struct {
	Content    string `json:"content,omitempty"`
	ScriptType string `json:"scriptType,omitempty"`
}

const (
	PROPERTY_GET = "propertyGet"
	PROPERTY_SET = "propertySet"
)

type IssueTcpCommandPl struct {
	DeviceId        int         `json:"deviceId,omitempty"`
	FunctionType    string      `json:"functionType,omitempty"`
	Identifier      string      `json:"identifier,omitempty"`
	IdentifierValue interface{} `json:"identifierValue,omitempty"`
}

type IssueMQTTCommandPl struct {
	DeviceId        int         `json:"deviceId,omitempty"`
	FunctionType    string      `json:"functionType,omitempty"`
	Identifier      string      `json:"identifier,omitempty"`
	IdentifierValue interface{} `json:"identifierValue,omitempty"`
}

const (
	ROUTER_ADDR_REST  = 1
	ROUTER_ADDR_MQTT  = 2
	ROUTER_ADDR_KAFKA = 3
	ROUTER_ADDR_MYSQL = 4
)

type CreateDMPRouterAddressPl struct {
	Name string `json:"name,omitempty"`
	Type int    `json:"type,omitempty"`
	Env  int    `json:"env,omitempty"`
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
	//Rest特有参数
	Path string `json:"path,omitempty"`
	//MQTT特有参数
	ClientId string `json:"clientId,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Topic    string `json:"topic,omitempty"`
	//KAFKA特有参数
	Addr string `json:"addr,omitempty"`
	//MYSQL特有参数
	Database string `json:"database,omitempty"`
	Table    string `json:"table,omitempty"`
}

const (
	ROUTER_TYPE_STANDARD     = 1
	ROUTER_TYPE_RULEENGINE   = 2
	ROUTER_TYPE_DATAANALYZE  = 3
	ROUTER_FORMAT_JSON       = 1
	ROUTER_FORMAT_XML        = 2
	ROUTER_FORMAT_SERIALIZED = 3
	ROUTER_FORMAT_CSV        = 4
	ROUTER_COMPRESSION_NONE  = 1
	ROUTER_COMPRESSOIN_GZIP  = 2
	ROUTER_COMPRESSION_ZIP   = 3
)

type CreateDMPRouterPl struct {
	Name            string `json:"name,omitempty"`
	Env             int    `json:"env,omitempty"`
	Type            int    `json:"type,omitempty"`
	AddressType     int    `json:"addressType,omitempty"`
	AddressConfigId int    `json:"addressConfigId,omitempty"`
	Format          int    `json:"format,omitempty"`
	Encryption      struct {
		EncryptionAlgorithm int    `json:"encryptionAlgorithm,omitempty"`
		EncryptionKey       string `json:"encryptionKey,omitempty"`
		InitializingVector  string `json:"initializingVector,omitempty"`
	} `json:"encryption,omitempty"`
	Compression int    `json:"compression,omitempty"`
	Filter      Filter `json:"filter,omitempty"`
}

type Filter struct {
	DevIdentifiers []DevIdentifier `json:"devIdentifiers,omitempty"`
}

type DevIdentifier struct {
	Pid                        string   `json:"pid,omitempty"`
	DeviceId                   string   `json:"deviceId,omitempty"`
	ValueDescriptorIdentifiers []string `json:"valueDescriptorIdentifiers,omitempty"`
}

const (
	MODBUS_ORIGIN_DATA_TYPE_INT16 = 4
	MODBUS_READ_HOLDING_REGISTER  = "0x03"
	MODBUS_WRITE_HOLDING_REGISTER = "0x06"
	MODBUS_SWAP_BYTE_DISABLED     = 0
	MODBUS_SWAP_BYTE_ENABLED      = 1
	MODBUS_SWAP_ORDER_DISABLED    = 0
	MODBUS_SWAP_ORDER_ENABLED     = 1
	MODBUS_REPORTMETHOD_TIME      = 1
)

type CreateModbusProperty struct {
	AccessMode      int     `json:"accessMode,omitempty"`
	Name            string  `json:"name,omitempty"`
	Identifier      string  `json:"identifier,omitempty"`
	Unit            string  `json:"unit,omitempty"`
	ReadFlag        string  `json:"readFlag,omitempty"`
	WriteFlag       string  `json:"writeFlag,omitempty"`
	Type            int     `json:"type,omitempty"`
	SwapByte        int     `json:"swapByte,omitempty"`
	SwapOrder       int     `json:"swapOrder,omitempty"`
	Scalingfactor   float64 `json:"scalingfactor,omitempty"`
	ReportMethod    int     `json:"reportMethod,omitempty"`
	RegisterAddress string  `json:"registerAddress,omitempty"`
	RegisterNumber  int     `json:"registerNumber,omitempty"`
	OriginDataType  int     `json:"originDataType,omitempty"`
	Minimum         float64 `json:"minimum"`
	Maximum         float64 `json:"maximum,omitempty"`
	Special         Special `json:"special,omitempty"`
}

const (
	LWM2M_PROPERTY_GET     = "propertyGet"
	LWM2M_PROPERTY_SET     = "propertySet"
	LWM2M_PROPERTY_EXECUTE = "propertyExecute"
)

type IssueLWM2MCommandPl struct {
	DeviceId        int         `json:"deviceId,omitempty"`
	FunctionType    string      `json:"functionType,omitempty"`
	Identifier      string      `json:"identifier,omitempty"`
	IdentifierValue interface{} `json:"identifierValue,omitempty"`
}

type AddModbusChannelToDTUPl struct {
	Id            int `json:"id,omitempty"`
	ModbusAddress int `json:"modbusAddress,omitempty"`
	CollectTime   int `json:"collectTime,omitempty"`
}

type IssueModbusCommandByNamePl struct {
	ProductId       int         `json:"productId,omitempty"`
	DeviceName      string      `json:"deviceName,omitempty"`
	FunctionType    string      `json:"functionType,omitempty"`
	Identifier      string      `json:"identifier,omitempty"`
	IdentifierValue interface{} `json:"identifierValue,omitempty"`
}

type CreateRulePl struct {
	Name           string          `json:"name,omitempty"`
	Conditions     []RuleCondition `json:"conditions,omitempty"`
	TimeTrigger    string          `json:"timeTrigger,omitempty"`
	Filters        []RuleFilter    `json:"filters,omitempty"`
	Actions        []RuleAction    `json:"actions,omitempty"`
	Subscription   []string        `json:"subscription,omitempty"`
	ExportClientId int             `json:"exportClientId,omitempty"`
	MailServerId   string          `json:"mailServerId,omitempty"`
	MailPatternId  string          `json:"mailPatternId,omitempty"`
	RuleType       int             `json:"ruleType,omitempty"`
}

type RuleCondition struct {
	ProductId int         `json:"productId,omitempty"`
	DidList   []int       `json:"didList,omitempty"`
	Parameter string      `json:"parameter,omitempty"`
	Operation string      `json:"operation,omitempty"`
	Operand   interface{} `json:"operand,omitempty"`
}

type RuleFilter struct {
	Type         int `json:"type,omitempty"`
	FilterDevice struct {
		ProductId int         `json:"productId,omitempty"`
		DidList   []int       `json:"didList,omitempty"`
		Parameter string      `json:"parameter,omitempty"`
		Operation string      `json:"operation,omitempty"`
		Operand   interface{} `json:"operand,omitempty"`
	} `json:"filterDevice,omitempty"`
	FilterTime struct {
		Start string `json:"start,omitempty"`
		End   string `json:"end,omitempty"`
	} `json:"filterTime,omitempty"`
}

type RuleAction struct {
	ProductId  int         `json:"productId,omitempty"`
	DeviceId   int         `json:"deviceId,omitempty"`
	Identifier string      `json:"identifier,omitempty"`
	Value      interface{} `json:"value,omitempty"`
}

type CreateMailServerPl struct {
	Name       string `json:"name"`
	Account    string `json:"account"`
	AuthCode   string `json:"authCode"`
	ServerHost string `json:"serverHost"`
	ServerPort string `json:"serverPort"`
	Type       int    `json:"type"`
}

type CreateMailTemplate struct {
	Name    string `json:"name"`
	Content string `json:"content"`
	Type    int    `json:"type"`
}
