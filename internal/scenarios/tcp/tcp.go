package tcp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"oes-acceptance-testing/v1/internal/tester"
	"oes-acceptance-testing/v1/pkg/helpers/conf"
	"oes-acceptance-testing/v1/pkg/helpers/logger"
	"oes-acceptance-testing/v1/pkg/helpers/random"
	"strconv"
	"sync"
	"time"
)

const (
	suiteName = "TCP透传测试套件"
	td        = true
)

type cmdResp struct {
	Uuid            string      `json:"uuid"`
	IdentifierValue interface{} `json:"identifierValue"`
}

type tcpCommand struct {
	Uuid         string
	FunctionType string
	Params       map[string]interface{}
}

var wg sync.WaitGroup

func RunTcpTest(resultChan chan tester.TestResult, onHold bool) {
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

	// 创建TCP产品
	cp := tester.CreateDmpProductPl{
		Name:                 "tcp_" + random.RandString(),
		Description:          "some description",
		ProtocolType:         3,
		NodeType:             1,
		Model:                1,
		DataFormat:           2,
		AuthenticationMethod: 1,
		NetworkMethod:        1,
	}
	t.Record, err = t.CreateDmpProduct(cp)
	t.Must(err)

	pInfo := t.Record.Response.(*tester.GeneralResp)
	t.AssertTrue("创建TCP产品", pInfo.Success)

	t.Params.ProductID = pInfo.Data.ID
	t.Params.MasterKey = pInfo.Data.MasterKey

	// 创建功能
	t.CreateProperties()

	//添加设备
	cd := tester.CreateDmpDevicePl{
		Name:        "some_device",
		Description: "some device",
	}

	t.Record, err = t.CreateDmpDevice(t.Params.ProductID, cd)
	t.Must(err)
	t.AssertTrue("添加设备", t.Record.IsSuccess())
	dInfo := t.Record.Response.(*tester.GeneralResp)
	t.Params.DeviceID = dInfo.Data.ID
	t.Params.ApiKey = dInfo.Data.ApiKey

	//添加解析脚本
	scriptBytes, readErr := ioutil.ReadFile("data/tcpscript")
	if readErr != nil {
		t.Logger.Fatalf("error reading file: %v", err)
	}

	cts := tester.CreateTcpScriptPl{
		Content:    string(scriptBytes),
		ScriptType: "groovy",
	}
	t.Record, err = t.CreateTcpScript(t.Params.ProductID, t.Params.MasterKey, cts)
	t.Must(err)
	t.AssertTrue("添加tcp解析脚本", t.Record.IsSuccess())

	// 设备连接及鉴权

	// 连接到平台
	conn, err := net.DialTimeout("tcp", conf.TcpServer, 10*time.Second)
	if err != nil {
		t.Logger.Printf("error connecting to tcp server: %v", err)
		panic(err)
	}

	// hash apikey
	token := encryptApiKey(strconv.Itoa(t.Params.ProductID), strconv.Itoa(t.Params.DeviceID), t.Params.ApiKey)

	// 监听并处理平台下发的命令
	go handleCommand(t, conn)

	// 发送注册报文
	authKey := fmt.Sprintf(`*%v#%v#%v*`, t.Params.ProductID, t.Params.DeviceID, token)
	_, err = conn.Write([]byte(authKey))
	t.Must(err)
	t.Logger.Printf("已发送鉴权信息: %v\n", authKey)

	// 等待1秒
	t.Logger.Println("等待1秒...")
	time.Sleep(time.Second * 2)

	// 上行数据
	go upwardData(t, conn)

	// 心跳
	go heartBeat(t, conn)

	// 等待3秒
	t.Logger.Println("等待3秒...")
	time.Sleep(time.Second * 3)

	// 下发读命令
	itc := tester.IssueTcpCommandPl{
		DeviceId:     t.Params.DeviceID,
		FunctionType: "propertyGet",
		Identifier:   "someDouble",
	}
	t.Record, err = t.IssueTcpCommand(itc)
	t.Must(err)
	cInfo := t.Record.Response.(*tester.CmdResp)
	dp1, ok := cInfo.Data.Value.(float64)
	if !ok {
		t.FailTest("数据点异常，测试失败")
		t.Logger.Println(t.Record.ResponseString)
	} else {
		t.AssertFloat64("下发读命令应能读取到上行的值", float64(999999.999999999), dp1)
	}

	// 下发写命令
	itwc := tester.IssueTcpCommandPl{
		DeviceId:        t.Params.DeviceID,
		FunctionType:    "propertySet",
		Identifier:      "someDouble",
		IdentifierValue: 555555.555555555,
	}
	t.Record, err = t.IssueTcpCommand(itwc)
	t.Must(err)
	t.AssertTrue("下发写命令", t.Record.IsSuccess())

	// 检查上行数据点应已入库
	t.Record, err = t.GetMultipleLatestDataPoints(t.Params.DeviceID, []string{"someDouble"})
	t.Must(err)
	dp := t.Record.Response.(*tester.DataPointResp)
	if len(dp.Data) < 1 {
		t.FailTest("获取最新数据点接口未获取到数据点!测试失败")
	} else {
		t.AssertFloat64("获取最新数据点的值应正确", 999999.999999999, dp.Data[0].Value.(float64))
	}

	// 检查上行数据点应入库(分页查询资源某时间段数据点)
	hour, err := time.ParseDuration("168h")
	hm := hour.Milliseconds()
	if err != nil {
		err := fmt.Errorf("error while parsing duration: %v", err)
		panic(err)
	}
	start := time.Now().Unix()*1000 - hm
	end := time.Now().Unix() * 1000
	t.Record, err = t.GetDataPointsInTimePeriod(t.Params.DeviceID, "someDouble", start, end, 1, 20)
	t.Must(err)
	data3 := t.Record.Response.(*tester.ContentDataPointResp).Data.Content
	t.AssertIntegerGreaterThan("检查上行数据点应入库(分页查询资源某时间段数据点)", 2, len(data3))

	if onHold == true {
		wg.Add(1)
		wg.Wait()
	}

	if err := conn.Close(); err != nil {
		panic(err)
	}
}

func tearDown(t *tester.Tester, resultChan chan tester.TestResult) {
	var err error
	//删除TCP产品
	t.Record, err = t.DeleteDmpProduct(t.Params.ProductID)
	t.Must(err)
	t.AssertTrue("删除TCP设备", t.Record.IsSuccess())

	//判断套件是否成功并打印结果
	t.DisplayResult(resultChan)
}

func heartBeat(t *tester.Tester, conn net.Conn) {
	for {
		_, err := conn.Write([]byte("^^^^"))
		if err != nil {
			err := fmt.Errorf("上传心跳时发生错误: %v", err)
			t.Logger.Println(err)
			return
		}
		time.Sleep(10 * time.Second)
	}
}

func upwardData(t *tester.Tester, conn net.Conn) {
	data := tester.NewDataModel()
	delete(data, "someBytes")

	for {
		for k, v := range data {
			lb := make([]byte, 2)
			json := `{"%v":%v}`
			if k == "someString" {
				json = `{"%v":"%v"}`
			}
			if k == "someBytes" {
				json = `"%v": "%v"`
			}
			var j string
			j = fmt.Sprintf(json, k, v.Value)
			if k == "someBytes" {
				vv := v.Value.([]byte)
				vvv := base64.StdEncoding.EncodeToString(vv)
				j = fmt.Sprintf(json, k, vvv)
			}
			l := len(j)
			binary.BigEndian.PutUint16(lb, uint16(l))
			dts := append(lb, []byte(j)...)
			fmt.Printf("%-30v\t\t", j)

			var byteArr string
			for _, b := range dts {
				byteArr += fmt.Sprintf("%v,", b)
			}
			fmt.Println(byteArr)
			_, err := conn.Write(dts)
			if err != nil {
				t.Logger.Println(err)
				return
			}
		}
		time.Sleep(1 * time.Second)
	}

}

func handleCommand(t *tester.Tester, conn net.Conn) {
	for {
		time.Sleep(time.Millisecond)
		request := make([]byte, 1024)
		l, err := conn.Read(request)
		t.Logger.Println("收到命令：\n" + string(request[:l]))

		if err == io.EOF {
			t.Logger.Fatalln("收到EOF，服务器主动断开了连接，结束程序。")
		}

		if err != nil {
			t.Logger.Printf("error while reading from tcp connection: %v", err)
			return
		}

		if request[:l][0] == 6 {
			t.Logger.Printf("已接到平台回复的上线成功报文: %v, 上线成功！\n", request[:l])
			continue
		}

		var cmd tcpCommand

		if err := json.Unmarshal(request[:l], &cmd); err != nil {
			log.Fatalln(err)
		}

		var resp []byte

		switch cmd.FunctionType {
		case "propertySet": // "set" means this is a write command
			resp = setResp(cmd)
			fmt.Println(resp)
			fmt.Println(string(resp))
			n, err := conn.Write(resp)
			t.Must(err)
			t.Logger.Printf("收到命令，发送回复: %v, 长度: %v\n", string(resp), n)

			var byteDisplay string
			for _, v := range resp {
				byteDisplay += fmt.Sprintf("%v", v) + ","
			}
			t.Logger.Printf("收到命令，发送回复: %v, 长度: %v\n", byteDisplay, n)

		}
	}
}

func setResp(cmd tcpCommand) []byte {
	startSymbol := []byte("#")
	type setCmdResp struct {
		Msg  string `json:"msg"`
		Uuid string `json:"uuid"`
	}
	resp := setCmdResp{
		Msg:  "success",
		Uuid: cmd.Uuid,
	}
	jString, err := json.Marshal(&resp)
	if err != nil {
		panic(fmt.Errorf("Error: generateResp: %v", err))
	}
	jStringLenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(jStringLenBytes, uint16(len(jString)))

	var payload []byte
	buffer := bytes.NewBuffer(payload)
	buffer.Write(startSymbol)
	buffer.Write(jStringLenBytes)
	buffer.Write(jString)

	return buffer.Bytes()
}

func encryptApiKey(pid, did, deviceSecret string) string {
	query := did + "&" + pid

	h := hmac.New(sha1.New, []byte(deviceSecret))
	h.Write([]byte(query))

	token := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return token
}
