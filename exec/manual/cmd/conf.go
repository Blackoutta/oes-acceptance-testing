package cmd

type conf struct {
	Pid                   int
	ProdSecret            string
	Did                   int
	DevSecret             string
	DevName               string
	Qos                   byte
	Interval              int
	Host                  string
	DynamicRegisterServer string
	PayloadFile           string
	ShadowFile            string
}

type lwm2mconf struct {
	Imei        string
	Interval    int
	Host        string
	PayloadFile string
}

type modbusconf struct {
	Pid       int
	Did       int
	DevSecret string
	SlaveID   byte
	Host      string
}
