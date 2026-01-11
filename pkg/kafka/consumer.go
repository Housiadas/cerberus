package kafka

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	MinCommitCount = 4
)

type Consumer interface {
	Subscribe(topic string) error
	Consume(ctx context.Context, msg *kafka.Message) error
	Close()
}

type ConsumerConfig struct {
	Brokers          string
	GroupID          string
	AddressFamily    string
	SecurityProtocol string
	SessionTimeout   int
}

type ConsumerClient struct {
	consumer *kafka.Consumer
	log      *logger.Service
}

func NewConsumer(log *logger.Service, cfg ConsumerConfig) (*ConsumerClient, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        cfg.Brokers,
		"group.id":                 cfg.GroupID,
		"broker.address.family":    cfg.AddressFamily,
		"session.timeout.ms":       cfg.SessionTimeout,
		"auto.offset.reset":        "earliest",
		"enable.auto.offset.store": false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &ConsumerClient{
		consumer: consumer,
		log:      log,
	}, nil
}

func (c *ConsumerClient) Close() {
	c.consumer.Close()
}

func (c *ConsumerClient) Subscribe(topic string) error {
	err := c.consumer.Subscribe(topic, nil)
	if err != nil {
		return fmt.Errorf("kafka subscribe error: %w", err)
	}

	return nil
}

func (c *ConsumerClient) Consume(ctx context.Context, fn func(msg *kafka.Message) error) error {
	msgCount := 0

	run := true
	for run {
		ev := c.consumer.Poll(100)
		switch event := ev.(type) {
		case *kafka.Message:
			msgCount++
			if msgCount%MinCommitCount == 0 {
				go func() {
					_, err := c.consumer.Commit()
					c.log.Error(ctx, fmt.Sprintf("consumer: Committing%v\n", err))
				}()
			}
			// Callback, application-specific
			err := fn(event)
			if err != nil {
				c.log.Error(ctx, fmt.Sprintf("consumer: %v\n", event))
			}

			fmt.Printf("%% Message on %s:\n%s\n", event.TopicPartition, string(event.Value))
		case kafka.PartitionEOF:
			c.log.Info(ctx, fmt.Sprintf("consumer: EOF Reached %v\n", event))
		case kafka.Error:
			c.log.Error(ctx, fmt.Sprintf("consumer: %v\n", event))

			run = false
		default:
			c.log.Info(ctx, fmt.Sprintf("consumer: Ignored %v\n", event))
		}
	}

	return nil
}
