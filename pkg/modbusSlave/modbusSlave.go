package modbusSlave

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"oes-acceptance-testing/v1/pkg/helpers/random"
	"strconv"
	"time"

	"github.com/sigurn/crc16"
)

var modbusCRCTable = crc16.MakeTable(crc16.CRC16_MODBUS)

type ModbusSlave struct {
	ProductID        int
	DeviceID         int
	DeviceSecret     string
	SlaveID          byte
	ModbusMasterAddr string
}

func (s ModbusSlave) Boot() {
	addr := s.ModbusMasterAddr

	log.Println("connecting...", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	log.Println("connected!")

	productID := strconv.Itoa(s.ProductID)
	deviceID := strconv.Itoa(s.DeviceID)
	deviceSecret := s.DeviceSecret

	// 鉴权
	token := encryptApiKey(productID, deviceID, deviceSecret)
	authKey := fmt.Sprintf("*%v#%v#%v*", productID, deviceID, token)

	log.Println("sending authkey:", authKey)
	_, err = conn.Write([]byte(authKey))
	if err != nil {
		panic(err)
	}

	// 判断及回复平台(modbus master)的命令
	go func() {
		for {
			time.Sleep(time.Millisecond)
			buf := make([]byte, 1024)
			l, err := conn.Read(buf)
			if err != nil {
				panic(err)
			}
			msg := buf[:l]
			log.Printf("received: %v", msg)

			bs, _ := ioutil.ReadFile("./data/payload/modbus.data")
			tmp, _ := strconv.Atoi(string(bs))
			temp := byte(tmp)

			humi := byte(random.RandInt(255))
			var data ModBusMsg
			switch msg[1] { //判断Function
			case 3: //READ HOLDING REGISTER
				switch msg[3] { //判断HOLDING REGISTER起始地址
				case 0: //0x0000 led_light
					data = ModBusMsg{s.SlaveID, 3, 2, 0, 0}.PayloadWithCRC16()
				case 1: // 0x0001 temp
					data = ModBusMsg{s.SlaveID, 3, 2, 0, temp}.PayloadWithCRC16()
					log.Printf("send temp: %v\n", temp)
				case 2: // 0x0002 humi
					data = ModBusMsg{s.SlaveID, 3, 4, 0, 0, 0, humi}.PayloadWithCRC16()
				case 3: // 0x0003 test
					data = ModBusMsg{s.SlaveID, 3, 8, 0, 0, 0, 0, 0, 0, 0, 255}.PayloadWithCRC16()
				}
			case 6: //PRESETS SINGLE REGISTER(WRITE HOLDING REGISTER)
				data = msg
			}

			_, err = conn.Write(data)
			if err != nil {
				panic(err)
			}
			log.Printf("send: %v\n", data)
		}
	}()

	// 心跳
	go func() {
		for {
			hb := []byte{0x23, 0x23}
			_, err = conn.Write(hb)
			if err != nil {
				panic(err)
			}
			log.Printf("已发送心跳: %v", hb)
			time.Sleep(20 * time.Second)
		}
	}()
}

type ModBusMsg []byte

func (m ModBusMsg) PayloadWithCRC16() []byte {
	crcBytes := make([]byte, 2)
	crc := crc16.Checksum(m, modbusCRCTable)
	binary.LittleEndian.PutUint16(crcBytes, crc)
	return append(m, crcBytes...)
}

func encryptApiKey(pid, did, apiKey string) string {
	query := did + "&" + pid

	h := hmac.New(sha1.New, []byte(apiKey))
	h.Write([]byte(query))

	token := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return token
}
