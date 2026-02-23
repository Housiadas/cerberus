// Package kafka is a client for Apache Kafka
package kafka

type Event struct {
	Topic string
	Data  []byte
}
