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
		"enable.auto.commit":       true,
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

// Consume This uses StoreMessage (which stores the offset for that message)
// and relies on the auto-committer (enable.auto.commit defaults to true in confluent-kafka-go).
// No goroutines, no race conditions, and the context is respected.
func (c *ConsumerClient) Consume(ctx context.Context, fn func(msg *kafka.Message) error) error {
	for {
		err := c.pollOnce(ctx, fn)
		if err != nil {
			return err
		}
	}
}

func (c *ConsumerClient) pollOnce(ctx context.Context, fn func(msg *kafka.Message) error) error {
	// check for context cancellation
	select {
	case <-ctx.Done():
		return fmt.Errorf("consumer: context done: %w", ctx.Err())
	default:
	}

	ev := c.consumer.Poll(100)
	switch event := ev.(type) {
	case nil:
		return nil
	case *kafka.Message:
		return c.handleMessage(ctx, event, fn)
	case kafka.Error:
		if event.IsFatal() {
			return fmt.Errorf("consumer: fatal Kafka error: %w", event)
		}

		c.log.Error(ctx, fmt.Sprintf("consumer: retriable error: %v", event))
	case kafka.PartitionEOF:
		c.log.Info(ctx, fmt.Sprintf("consumer: partition EOF: %v", event))
	}

	return nil
}

func (c *ConsumerClient) handleMessage(
	ctx context.Context,
	event *kafka.Message,
	fn func(msg *kafka.Message) error,
) error {
	err := fn(event)
	if err != nil {
		c.log.Error(ctx, fmt.Sprintf("consumer: handler error: %v", err))

		return nil
	}

	_, storeErr := c.consumer.StoreMessage(event)
	if storeErr != nil {
		c.log.Error(ctx, fmt.Sprintf("consumer: store offset error: %v", storeErr))
	}

	return nil
}
