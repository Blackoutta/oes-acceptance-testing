package datarouter

import (
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// NewMQTTClient creates a new mqtt client using the values passed in
func NewMQTTClient(clientID, username, password, addr string, handler MQTT.MessageHandler) MQTT.Client {
	opts := MQTT.NewClientOptions().AddBroker(addr)
	opts.SetClientID(clientID)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetCleanSession(true)
	opts.SetKeepAlive(35 * time.Second)
	opts.SetDefaultPublishHandler(handler)
	return MQTT.NewClient(opts)
}
