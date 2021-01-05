package mqttClient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type EdgeMqttClient struct {
	ServerHostAndPort string
	ProductID         string
	DeviceKey         string
	DeviceName        string
	DeviceID          string
	PemFile           string
	AuthMsg           AuthMessage
	MQTT.Client
}

// TLS Option
func AddTLS(opts *MQTT.ClientOptions, pemFile string) {
	opts.SetTLSConfig(NewTLSConfig(pemFile))
}

func (e *EdgeMqttClient) NewMqttClient(keepAlive float64, handler MQTT.MessageHandler) error {
	opts := MQTT.NewClientOptions()
	if e.PemFile != "" {
		AddTLS(opts, e.PemFile)
	}
	opts.AddBroker(e.ServerHostAndPort)
	opts.SetClientID(e.DeviceID)
	opts.SetUsername(e.ProductID)
	password, err := GenerateSasToken(e.AuthMsg, e.ProductID, e.DeviceID, e.DeviceKey)
	if err != nil {
		return err
	}
	opts.SetPassword(string(password))
	opts.SetKeepAlive(time.Duration(keepAlive) * time.Second)
	opts.SetProtocolVersion(4)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(false)
	opts.SetDefaultPublishHandler(handler)
	e.Client = MQTT.NewClient(opts)
	fmt.Println("username:", e.ProductID)
	fmt.Println("password:", string(password))
	fmt.Println("host:", e.ServerHostAndPort)
	fmt.Println("device key", e.DeviceKey)
	return nil
}

func NewTLSConfig(pemFile string) *tls.Config {
	certPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(pemFile)
	if err != nil {
		panic(err)
	}
	certPool.AppendCertsFromPEM(pem)
	return &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: true,
	}
}
