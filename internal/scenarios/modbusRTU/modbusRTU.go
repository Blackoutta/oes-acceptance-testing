package modbusRTU

import (
	"fmt"
	"net/http"
	"oes-acceptance-testing/v1/internal/tester"
	"oes-acceptance-testing/v1/pkg/helpers/conf"
	"oes-acceptance-testing/v1/pkg/helpers/logger"
	"oes-acceptance-testing/v1/pkg/helpers/random"
	"oes-acceptance-testing/v1/pkg/modbusSlave"
	"strconv"
	"sync"
	"time"
)

const (
	suiteName = "MODBUS RTU测试套件"
	td        = true
)

var wg sync.WaitGroup

func RunModbusRTUTest(resultChan chan tester.TestResult, onHold bool) {
	var err error
	// initialize logger with suite name
	le, f := logger.NewLogger(suiteName)
	defer f.Close()

	// initialize tester
	t := &tester.Tester{
		SuiteName: suiteName,
		Logger:    le,
		Teardown:  td,
		Client:    &http.Client{},
	}

	// set up teardown
	if t.Teardown == true {
		defer tearDown(t, resultChan)
	}

	// recover if panic, then proceed to teardown
	defer t.Recover()

	// 创建DTU产品
	cp := tester.CreateDmpProductPl{
		Name:                 "dtu_" + random.RandString(),
		Description:          "some description",
		ProtocolType:         tester.PROTOCOL_TCP,
		NodeType:             tester.NODETYPE_DTU,
		Model:                tester.MODEL_NONE,
		DataFormat:           tester.DATAFORMAT_MODBUSRTU,
		AuthenticationMethod: tester.AUTHMETHOD_DEVICEKEY,
		NetworkMethod:        1,
	}

	t.Record, err = t.CreateDmpProduct(cp)
	t.Must(err)
	t.AssertTrue("创建DTU产品", t.Record.IsSuccess())
	pInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.ProductID = pInfo.Data.ID
	t.Params.MasterKey = pInfo.Data.MasterKey

	//添加DTU设备
	cd := tester.CreateDmpDevicePl{
		Name:        "dtu_device",
		Description: "some device",
	}

	t.Record, err = t.CreateDmpDevice(t.Params.ProductID, cd)
	t.Must(err)
	t.AssertTrue("添加设备", t.Record.IsSuccess())
	dInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.DeviceID = dInfo.Data.ID
	t.Params.ApiKey = dInfo.Data.ApiKey
	fmt.Println(t.Params.ApiKey)

	// 创建MODBUS产品
	modbusProd := tester.CreateDmpProductPl{
		Name:                 "modbus_" + random.RandString(),
		Description:          "some description",
		ProtocolType:         tester.PROTOCOL_MODBUS,
		NodeType:             tester.NODETYPE_DEVICE,
		Model:                tester.MODEL_OBJECTMODEL,
		DataFormat:           tester.DATAFORMAT_MODBUSRTU,
		AuthenticationMethod: tester.AUTHMETHOD_DEVICEKEY,
		NetworkMethod:        1,
	}

	t.Record, err = t.CreateDmpProduct(modbusProd)
	t.Must(err)
	t.AssertTrue("创建Modbus产品", t.Record.IsSuccess())
	mInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.ModbusProdID = mInfo.Data.ID
	t.Params.ModbusMasterKey = mInfo.Data.MasterKey

	// 创建Modbus设备
	md := tester.CreateDmpDevicePl{
		Name:        "modbus_device",
		Description: "some device",
	}

	t.Record, err = t.CreateDmpDevice(t.Params.ModbusProdID, md)
	t.Must(err)
	t.AssertTrue("添加Modbus设备", t.Record.IsSuccess())
	mdInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.ModbusDeviceID = mdInfo.Data.ID
	t.Params.ModbusApiKey = mdInfo.Data.ApiKey

	// 创建Modbus功能
	// temperature 0x03 | 0x0001
	temp := tester.CreateModbusProperty{
		AccessMode: 1,
		Name:       "温度监测",
		Identifier: "temperature",
		Unit:       "摄氏度",
		ReadFlag:   tester.MODBUS_READ_HOLDING_REGISTER,
		WriteFlag:  "",
		Type:       tester.DATATYPE_FLOAT64,
		Minimum:    -999999.999999999,
		Maximum:    999999.999999999,
		Special: tester.Special{
			Step: 0.000000001,
		},
		SwapByte:        tester.MODBUS_SWAP_BYTE_DISABLED,
		SwapOrder:       tester.MODBUS_SWAP_ORDER_DISABLED,
		Scalingfactor:   0.1,
		ReportMethod:    tester.MODBUS_REPORTMETHOD_TIME,
		RegisterAddress: "0x0001",
		RegisterNumber:  0,
		OriginDataType:  tester.MODBUS_ORIGIN_DATA_TYPE_INT16,
	}

	t.Record, err = t.CreateModbusProperty(temp, t.Params.ModbusProdID)
	t.Must(err)
	t.AssertTrue("添加温度功能(只读)", t.Record.IsSuccess())

	humi := tester.CreateModbusProperty{
		AccessMode: 1,
		Name:       "湿度监测",
		Identifier: "humidity",
		Unit:       "百帕",
		ReadFlag:   tester.MODBUS_READ_HOLDING_REGISTER,
		WriteFlag:  "",
		Type:       tester.DATATYPE_FLOAT64,
		Minimum:    -999999.999999999,
		Maximum:    999999.999999999,
		Special: tester.Special{
			Step: 0.000000001,
		},
		SwapByte:        tester.MODBUS_SWAP_BYTE_DISABLED,
		SwapOrder:       tester.MODBUS_SWAP_ORDER_DISABLED,
		Scalingfactor:   0.1,
		ReportMethod:    tester.MODBUS_REPORTMETHOD_TIME,
		RegisterAddress: "0x0002",
		RegisterNumber:  0,
		OriginDataType:  5,
	}

	t.Record, err = t.CreateModbusProperty(humi, t.Params.ModbusProdID)
	t.Must(err)
	t.AssertTrue("添加湿度功能(只读)", t.Record.IsSuccess())

	test := tester.CreateModbusProperty{
		AccessMode: 1,
		Name:       "test",
		Identifier: "test",
		Unit:       "test",
		ReadFlag:   tester.MODBUS_READ_HOLDING_REGISTER,
		WriteFlag:  "",
		Type:       tester.DATATYPE_FLOAT64,
		Minimum:    -999999.999999999,
		Maximum:    999999.999999999,
		Special: tester.Special{
			Step: 0.000000001,
		},
		SwapByte:        tester.MODBUS_SWAP_BYTE_DISABLED,
		SwapOrder:       tester.MODBUS_SWAP_ORDER_DISABLED,
		Scalingfactor:   0.1,
		ReportMethod:    tester.MODBUS_REPORTMETHOD_TIME,
		RegisterAddress: "0x0003",
		OriginDataType:  6,
	}

	t.Record, err = t.CreateModbusProperty(test, t.Params.ModbusProdID)
	t.Must(err)
	t.AssertTrue("添加测试功能", t.Record.IsSuccess())

	// led_light 0x03 0x06 | 0x000
	light := tester.CreateModbusProperty{
		AccessMode:      1,
		Name:            "LED灯",
		Identifier:      "led_light",
		Unit:            "",
		ReadFlag:        tester.MODBUS_READ_HOLDING_REGISTER,
		WriteFlag:       tester.MODBUS_WRITE_HOLDING_REGISTER,
		Type:            tester.DATATYPE_INT32,
		Minimum:         0,
		Maximum:         1,
		SwapByte:        tester.MODBUS_SWAP_BYTE_DISABLED,
		SwapOrder:       tester.MODBUS_SWAP_ORDER_DISABLED,
		Scalingfactor:   1,
		ReportMethod:    tester.MODBUS_REPORTMETHOD_TIME,
		RegisterAddress: "0x0000",
		RegisterNumber:  0,
		OriginDataType:  tester.MODBUS_ORIGIN_DATA_TYPE_INT16,
		Special: tester.Special{
			Step: 1,
		},
	}
	t.Record, err = t.CreateModbusProperty(light, t.Params.ModbusProdID)
	t.Must(err)
	t.AssertTrue("添加LED灯功能(可读可写)", t.Record.IsSuccess())

	// 在DTU中关联Modbus设备
	t.Record, err = t.AssociateDTUWithModbusDevice(nil, t.Params.DeviceID, t.Params.ApiKey, t.Params.ModbusProdID, t.Params.ModbusMasterKey, strconv.Itoa(t.Params.ModbusDeviceID))
	t.Must(err)
	t.AssertTrue("DTU关联Modbus设备", t.Record.IsSuccess())

	// 配置DTU的Modbus通道
	modbusChannel := tester.AddModbusChannelToDTUPl{
		Id:            t.Params.ModbusDeviceID,
		ModbusAddress: 1,
		CollectTime:   10,
	}
	t.Record, err = t.AddModbusChannelToDTU(modbusChannel)
	t.Must(err)
	t.AssertTrue("配置DTU的Modbus通道", t.Record.IsSuccess())

	time.Sleep(2 * time.Second)

	// 设备上线
	slave1 := modbusSlave.ModbusSlave{
		ProductID:        t.Params.ProductID,
		DeviceID:         t.Params.DeviceID,
		DeviceSecret:     t.Params.ApiKey,
		SlaveID:          1,
		ModbusMasterAddr: conf.ModbusRTUServer,
	}

	go func() {
		slave1.Boot()
	}()

	time.Sleep(10 * time.Second)
	// 下发读命令读取temperature
	readCmd := tester.IssueModbusCommandByNamePl{
		ProductId:    t.Params.ModbusProdID,
		DeviceName:   "modbus_device",
		FunctionType: tester.PROPERTY_GET,
		Identifier:   "temperature",
	}
	t.Record, err = t.IssueModbusDTUCommandByName(readCmd)
	t.Must(err)
	t.AssertTrue("根据产品ID和设备名读取temperature", t.Record.IsSuccess())

	rdCmdId := t.Record.Response.(*tester.IssueAsyncCmdResp).Data.CommandId
	t.Logger.Println("read command id is:", rdCmdId)

	// 下发写命令修改led_light
	writeCmd := tester.IssueModbusCommandByNamePl{
		ProductId:       t.Params.ModbusProdID,
		DeviceName:      "modbus_device",
		FunctionType:    tester.PROPERTY_SET,
		Identifier:      "led_light",
		IdentifierValue: 1,
	}
	t.Record, err = t.IssueModbusDTUCommandByName(writeCmd)
	t.Must(err)
	t.AssertTrue("根据产品ID和设备名写入led_light", t.Record.IsSuccess())

	wrtCmdId := t.Record.Response.(*tester.IssueAsyncCmdResp).Data.CommandId
	t.Logger.Println("write command id is:", wrtCmdId)

	t.Logger.Println("等待5秒后获取异步命令结果")
	time.Sleep(5 * time.Second)

	// 获取读命令结果
	t.Record, err = t.GetAsyncCommandResult(nil, rdCmdId)
	t.Must(err)
	t.AssertTrue("获取读命令结果", t.Record.IsSuccess())
	tempResp := t.Record.Response.(*tester.GetAsyncCmdResp)
	tempVal, ok := tempResp.Data.Value.(float64)
	if !ok {
		t.Logger.Println("temp数据点读取异常 ....fail")
	}
	tempResult := tempResp.Data.Result
	t.AssertFloat64("temperature的值等于23.5", 23.5, tempVal)
	t.AssertInteger("result的值等于3", 3, tempResult)

	// 获取写命令结果
	t.Record, err = t.GetAsyncCommandResult(nil, wrtCmdId)
	t.Must(err)
	t.AssertTrue("获取写命令结果", t.Record.IsSuccess())
	ledResp := t.Record.Response.(*tester.GetAsyncCmdResp)
	ledVal, ok := ledResp.Data.Value.(float64)
	if !ok {
		t.Logger.Println("写命令数据点异常! ... fail")
	}
	ledResult := ledResp.Data.Result
	t.AssertFloat64("led的值为1", 1, ledVal)
	t.AssertInteger("result的值等于3", 3, ledResult)

	// 获取temperature数据点
	t.Record, err = t.GetMultipleLatestDataPoints(t.Params.ModbusDeviceID, []string{"temperature"})
	t.Must(err)
	t.AssertTrue("获取modbus dtu数据点可以成功", t.Record.IsSuccess())
	data := t.Record.Response.(*tester.DataPointResp).Data
	if len(data) < 1 {
		err := fmt.Errorf("未获取到数据点，测试失败")
		t.Logger.Println(t.Record.ResponseString)
		panic(err)
	} else {
		t.AssertFloat64("温度数据点应等于23.5(精确到小数点后一位)", 23.5, data[0].Value.(float64))
		t.Logger.Println(t.Record.ResponseString)
	}

	// set up onhold
	if onHold == true {
		wg.Add(1)
		wg.Wait()
	}

}

func tearDown(t *tester.Tester, resultChan chan tester.TestResult) {
	//删除DTU产品
	var err error
	t.Record, err = t.DeleteDmpProduct(t.Params.ProductID)
	t.Must(err)
	t.AssertTrue("删除DTU产品", t.Record.IsSuccess())

	//删除MODBUS
	t.Record, err = t.DeleteDmpProduct(t.Params.ModbusProdID)
	t.Must(err)
	t.AssertTrue("删除MODBUS产品", t.Record.IsSuccess())

	//判断套件是否成功并打印结果
	t.DisplayResult(resultChan)
}
