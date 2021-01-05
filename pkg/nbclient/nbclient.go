package nbclient

import (
	"fmt"
	"log"
	"oes-acceptance-testing/v1/pkg/helpers/random"
	"oes-acceptance-testing/v1/pkg/tlv"
	"strconv"
	"time"

	"github.com/dustin/go-coap"
)

type NbiotClient struct {
	Addr     string
	IMEI     string
	IMSI     string
	Interval time.Duration
}

func (nc *NbiotClient) Boot() {
	if nc.Interval == 0 {
		nc.Interval = 5
	}
	temp := &temperature{
		sensorValue: 23.5,
	}

	sp := &setPoint{
		setPointValue: 77.777,
	}

	resourceMap := make(map[string]string)
	resourceMap["3303"] = `</3303/0/5605>,</3303/0/5700>`
	resourceMap["3308"] = `</3308/0/5900>`

	// =================与接入机建立连接===============//
	log.Printf("正在连接接入机: %v", nc.Addr)
	c, err := coap.Dial("udp", nc.Addr)
	if err != nil {
		log.Printf("error while dialing: %v", err)
		return
	}
	log.Println("连接成功！")

	// =================== 发送鉴权请求================= //
	log.Println("正在发送鉴权请求...")
	register := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.POST,
		MessageID: 65535,
		Payload:   []byte(`</>;rt="oma.lwm2m",</3308/0>,</3303/0>`),
	}
	register.SetPathString("/rd")                                         //请求路径
	register.AddOption(coap.ContentFormat, coap.AppLinkFormat)            //contentformat，相当于http的Content-Type. LinkFormat多用于资源发现
	register.AddOption(coap.URIQuery, fmt.Sprintf("ep=%v;imsi", nc.IMEI)) //鉴权信息，格式ep=imei;imsi
	register.AddOption(coap.URIQuery, "b=U")                              //设置字符集
	register.AddOption(coap.URIQuery, "lt=300")                           //保活时间
	fmt.Println(nc.IMEI)

	rv, err := c.Send(register)
	if err != nil {
		log.Printf("error while sending: %v", err)
		return
	}

	log.Printf("鉴权请求发送成功，服务器返回消息，Type: %v, 响应码:%v", rv.Type, rv.Code)
	log.Printf("鉴权成功！")

	// =================== 监听平台请求及命令================= //
	for {
		rv, err := c.Receive()
		if rv == nil {
			continue
		}
		if err != nil {
			log.Fatalf("error while receiving: %v", err)
		}

		// 如果收到ACK EMPTY MESSAGE，重新获取
		if a := rv.Code; a == 0 {
			log.Println("→ 收到Notify ACK")
			continue
		}

		// 如果收到的是Discover请求(比如:GET /3200)，则回复资源节点清单。payload中直接以byte stream形式回复字符串的节点清单，格式类型为："</3200/0/5500>,</3200/0/5750>"
		d, ok := rv.Option(coap.Accept).(coap.MediaType)
		if ok {
			if d == coap.AppLinkFormat {
				log.Printf("→ 收到Discover请求, MSG ID: %v", rv.MessageID)
				uripath := rv.Option(coap.URIPath).(string)
				accept := rv.Option(coap.Accept)
				payload := resourceMap[uripath]
				ack := coap.Message{
					Type:      coap.Acknowledgement,
					Code:      coap.Content,
					MessageID: rv.MessageID,
					Token:     rv.Token,
					Payload:   []byte(payload),
				}
				ack.SetPath(rv.Path())
				ack.AddOption(coap.ContentFormat, accept)

				_, err := c.Send(ack)
				if err != nil {
					log.Fatalf("error sending ack for Resouce List Acquisition: %v", err)
				}
				log.Printf("← 已回复Discover请求，MSG ID：%v", rv.MessageID)
				continue //回复成功后重新接收下一条数据
			}
		}

		// 如果收到的是Observe请求(比如: GET /3200/0)，回复所有节点中的值，payload中使用lwm2m tlv格式封装好各节点的值。并且开启持续notify。
		if o := rv.Option(coap.Observe); o != nil {

			uriPathSlice := rv.Path()

			resourceID, err := strconv.Atoi(uriPathSlice[0])
			if err != nil {
				panic(err)
			}

			var rs resource
			var identifier tlv.IntegerVal

			switch resourceID {
			case RESOURCE_TEMPERATURE:
				rs = temp
				identifier = TEMPERATURE_SENSOR_VALUE
			case RESOURCE_SETPOINT:
				rs = sp
				identifier = SETPOINT_SETPOINT_VALUE
			}

			payload := rs.getFloatValue().GeneratePayload(identifier)

			log.Printf("→ 收到Observe请求, MSG ID: %v", rv.MessageID)

			notifyObserve := 1

			// 针对Observe请求回复Ack
			accept := rv.Option(coap.Accept)
			ack := coap.Message{
				Type:      coap.Acknowledgement,
				Code:      coap.Content,
				MessageID: rv.MessageID,
				Token:     rv.Token,
				Payload:   payload, //TODO
			}
			ack.SetPath(rv.Path())
			ack.AddOption(coap.Observe, notifyObserve)
			ack.AddOption(coap.ContentFormat, accept)

			_, err = c.Send(ack)
			if err != nil {
				log.Fatalf("error sending Ack for Observe: %v", err)
			}
			log.Printf("← 已回复Observe请求，MSG ID：%v", rv.MessageID)

			//==========回复Observe后开始定期Notify==========//
			go func() {
				time.Sleep(2 * time.Second)
				for {
					payload := rs.getFloatValue().GeneratePayload(identifier)
					notifyObserve++
					notify := coap.Message{
						Type:      coap.Confirmable,
						Code:      coap.Content,
						MessageID: uint16(random.RandInt(99999)),
						Token:     rv.Token,
						Payload:   payload,
					}
					notify.SetPath(rv.Path())
					notify.AddOption(coap.Observe, notifyObserve)
					notify.AddOption(coap.ContentFormat, accept)
					err := c.Transmit(notify)
					if err != nil {
						log.Fatalf("error sending notify for Observe: %v", err)
					}
					log.Printf("← 已发送Notify，MSG ID：%v, TYPE: %v", notify.MessageID, notify.Type)

					time.Sleep(nc.Interval * time.Second)
				}
			}()
			continue
		}

		// 如果收到的是主动读取请求（比如: GET /3200/0/5500）
		if r := rv.Options(coap.URIPath); len(r) == 3 && rv.Code == coap.GET {
			log.Printf("→ 收到读命令(GET), URI PATH: %v, MSG ID: %v", rv.Options(coap.URIPath), rv.MessageID)

			accept := rv.Option(coap.Accept)

			uriPathSlice := rv.Path()

			resourceID, err := strconv.Atoi(uriPathSlice[0])
			if err != nil {
				panic(err)
			}
			identifierInt, err := strconv.Atoi(uriPathSlice[2])
			if err != nil {
				panic(err)
			}

			var identifier tlv.IntegerVal = tlv.IntegerVal(identifierInt)
			var payload []byte

			switch resourceID {
			case RESOURCE_TEMPERATURE:
				switch identifier {
				case TEMPERATURE_SENSOR_VALUE:
					payload = temp.sensorValue.GeneratePayload(identifier)
				}
			case RESOURCE_SETPOINT:
				switch identifier {
				case SETPOINT_SETPOINT_VALUE:
					payload = sp.setPointValue.GeneratePayload(identifier)
				}
			}

			content := coap.Message{
				Type:      coap.Acknowledgement,
				Code:      coap.Content,
				MessageID: rv.MessageID,
				Token:     rv.Token,
				Payload:   []byte(payload),
			}
			content.SetPath(uriPathSlice)
			content.AddOption(coap.ContentFormat, accept)

			_, err = c.Send(content)
			if err != nil {
				log.Fatalf("error sending ack for Resouce List Acquisition: %v", err)
			}
			log.Printf("← 已回复读命令(GET)请求，MSG ID：%v", content.MessageID)
			continue //回复成功后重新接收下一条数据
		}

		// 如果收到的是主动写入请求（比如：PUT /3200/0/5500）
		if rv.Code == coap.PUT {
			log.Printf("→ 收到写命令(PUT), URI PATH: %v, MSG ID: %v", rv.Options(coap.URIPath), rv.MessageID)
			changed := coap.Message{
				Type:      coap.Acknowledgement,
				Code:      coap.Changed,
				MessageID: rv.MessageID,
				Token:     rv.Token,
				Payload:   nil,
			}
			changed.SetPath(rv.Path())
			log.Println("put payload:", rv.Payload)

			_, err := c.Send(changed)
			if err != nil {
				log.Fatalf("error sending ack for Resouce List Acquisition: %v", err)
			}
			log.Printf("← 已回复写命令(PUT)请求，MSG ID：%v", changed.MessageID)
			continue //回复成功后重新接收下一条数据
		}

		if rv.Code == coap.POST {
			log.Printf("→ 收到执行(POST), URI PATH: %v, MSG ID: %v", rv.Options(coap.URIPath), rv.MessageID)
			changed := coap.Message{
				Type:      coap.Acknowledgement,
				Code:      coap.Changed,
				MessageID: rv.MessageID,
				Token:     rv.Token,
				Payload:   nil,
			}
			changed.SetPath(rv.Path())
			log.Println("post payload:", rv.Payload)

			_, err := c.Send(changed)
			if err != nil {
				log.Fatalf("error sending ack for Resouce List Acquisition: %v", err)
			}
			log.Printf("← 已回复执行命令(POST)请求，MSG ID：%v", changed.MessageID)
			continue //回复成功后重新接收下一条数据
		}
	}
}
