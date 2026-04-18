// Package kafka is a client for Apache Kafka
package kafka

import ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"

type Event struct {
	Topic string
	Data  []byte
}

func ExtractHeader(msg *ckafka.Message, key string) string {
	for _, h := range msg.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}

	return ""
}
