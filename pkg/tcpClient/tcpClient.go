package tcpClient

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
	"strconv"
	"time"
)

type TCPClient struct {
	ProductID int
	DeviceID  int
	DevSecret string
	conn      net.Conn
}

func (c *TCPClient) Connect(addr string) error {
	fmt.Println("dialing: ", addr)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return err
	}
	fmt.Println("dialed!")
	token := encryptApiKey(strconv.Itoa(c.ProductID), strconv.Itoa(c.DeviceID), c.DevSecret)
	authKey := fmt.Sprintf(`*%v#%v#%v*`, c.ProductID, c.DeviceID, token)
	fmt.Println("sending auth request...")
	_, err = conn.Write([]byte(authKey))
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *TCPClient) HandleCommand() {
	for {
		time.Sleep(time.Millisecond)
		request := make([]byte, 1024)
		l, err := c.conn.Read(request)
		log.Println("收到命令：\n" + string(request[:l]))

		if err == io.EOF {
			log.Fatalln("收到EOF，服务器主动断开了连接，结束程序。")
		}

		if err != nil {
			log.Printf("error while reading from tcp connection: %v", err)
			return
		}

		if request[:l][0] == 6 {
			log.Printf("已接到平台回复的上线成功报文: %v, 上线成功！\n", request[:l])
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
			n, err := c.conn.Write(resp)
			if err != nil {
				log.Println(err)
				time.Sleep(time.Second)
				continue
			}
			log.Printf("收到命令，发送回复: %v, 长度: %v\n", string(resp), n)

			var byteDisplay string
			for _, v := range resp {
				byteDisplay += fmt.Sprintf("%v", v) + ","
			}
			log.Printf("收到命令，发送回复: %v, 长度: %v\n", byteDisplay, n)

		}
	}
}

func (c *TCPClient) UpwardData(payloadFile string, interval time.Duration) {
	for {
		lb := make([]byte, 2)
		j, err := ioutil.ReadFile(payloadFile)
		if err != nil {
			if err != nil {
				log.Fatalf("读取payload文件时发生错误: %v", err)
			}
		}
		l := len(j)
		binary.BigEndian.PutUint16(lb, uint16(l))
		dts := append(lb, []byte(j)...)

		var byteArr string
		for _, b := range dts {
			byteArr += fmt.Sprintf("%v,", b)
		}
		// fmt.Println(byteArr)
		_, err = c.conn.Write(dts)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("已上行数据: %v\n", string(j))
		time.Sleep(interval * time.Second)
	}

}

func (c *TCPClient) HeartBeat() {
	for {
		_, err := c.conn.Write([]byte("^^^^"))
		if err != nil {
			err := fmt.Errorf("上传心跳时发生错误: %v", err)
			log.Println(err)
			return
		}
		time.Sleep(10 * time.Second)
	}
}

func encryptApiKey(pid, did, deviceSecret string) string {
	query := did + "&" + pid

	h := hmac.New(sha1.New, []byte(deviceSecret))
	h.Write([]byte(query))

	token := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return token
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
