package datarouter

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func NewKafkaConsumer(bootstrapServer string, topic string) (*kafka.Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServer,
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}
	c.SubscribeTopics([]string{topic, "^aRegex.*[Tt]opic"}, nil)
	return c, nil
}
