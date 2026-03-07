package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	defaultWorkers      = 4
	defaultBatchSize    = 100
	defaultFlushTimeout = 5 * time.Second
)

// Handler processes one message (called inside a worker goroutine).
type Handler func(ctx context.Context, msg *kafka.Message) error

// Flusher is called with a batch of successfully processed messages.
type Flusher func(ctx context.Context, msgs []*kafka.Message) error

// Consumer defines the interface for a Kafka consumer.
type Consumer interface {
	Subscribe(topic string) error
	Consume(ctx context.Context, handler Handler, flusher Flusher) error
	Close()
}

// ConsumerConfig holds configuration for ConsumerClient.
type ConsumerConfig struct {
	Brokers          string
	GroupID          string
	AddressFamily    string
	SecurityProtocol string
	SessionTimeout   int
	Workers          int
	BatchSize        int
	FlushTimeout     time.Duration
	BufferSize       int
}

// ConsumerClient is a Kafka consumer with a worker pool and batcher.
type ConsumerClient struct {
	consumer     *kafka.Consumer
	log          *logger.Service
	workers      int
	batchSize    int
	flushTimeout time.Duration
	bufferSize   int
}

// NewConsumer creates a new ConsumerClient.
func NewConsumer(log *logger.Service, cfg ConsumerConfig) (*ConsumerClient, error) {
	if cfg.Workers <= 0 {
		cfg.Workers = defaultWorkers
	}

	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultBatchSize
	}

	if cfg.FlushTimeout <= 0 {
		cfg.FlushTimeout = defaultFlushTimeout
	}

	if cfg.BufferSize <= 0 {
		cfg.BufferSize = cfg.Workers * cfg.BatchSize
	}

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
		consumer:     consumer,
		log:          log,
		workers:      cfg.Workers,
		batchSize:    cfg.BatchSize,
		flushTimeout: cfg.FlushTimeout,
		bufferSize:   cfg.BufferSize,
	}, nil
}

// Close shuts down the underlying Kafka consumer.
func (c *ConsumerClient) Close() {
	c.consumer.Close()
}

// Subscribe subscribes to the given Kafka topic.
func (c *ConsumerClient) Subscribe(topic string) error {
	err := c.consumer.Subscribe(topic, nil)
	if err != nil {
		return fmt.Errorf("kafka subscribe error: %w", err)
	}

	return nil
}

// Consume starts the worker pool and batcher.
// It blocks until the context is canceled or a fatal error occurs.
func (c *ConsumerClient) Consume(
	ctx context.Context,
	handler Handler,
	flusher Flusher,
) error {
	msgCh := make(chan *kafka.Message, c.bufferSize)
	processedCh := make(chan *kafka.Message, c.bufferSize)

	var wg sync.WaitGroup
	for range c.workers {
		wg.Go(func() {
			c.runWorker(ctx, msgCh, processedCh, handler)
		})
	}

	batcher := NewBatcher(c.consumer, flusher, c.batchSize, c.flushTimeout, c.log)

	batchErrCh := make(chan error, 1)

	go func() {
		batchErrCh <- batcher.Run(ctx, processedCh)
	}()

	pollErr := c.runPoller(ctx, msgCh)
	close(msgCh)

	wg.Wait()
	close(processedCh)

	batchErr := <-batchErrCh

	if pollErr != nil {
		return pollErr
	}

	return batchErr
}

func (c *ConsumerClient) runPoller(
	ctx context.Context,
	msgCh chan<- *kafka.Message,
) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("consumer: context done: %w", ctx.Err())
		default:
		}

		err := c.handleEvent(ctx, c.consumer.Poll(100), msgCh)
		if err != nil {
			return err
		}
	}
}

func (c *ConsumerClient) handleEvent(
	ctx context.Context,
	ev kafka.Event,
	msgCh chan<- *kafka.Message,
) error {
	switch event := ev.(type) {
	case nil:
		// no event
	case *kafka.Message:
		select {
		case msgCh <- event:
		case <-ctx.Done():
			return fmt.Errorf("consumer: context done: %w", ctx.Err())
		}
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

func (c *ConsumerClient) runWorker(
	ctx context.Context,
	msgCh <-chan *kafka.Message,
	processedCh chan<- *kafka.Message,
	handler Handler,
) {
	for msg := range msgCh {
		err := handler(ctx, msg)
		if err != nil {
			c.log.Error(ctx, fmt.Sprintf("consumer: handler error: %v", err))

			continue
		}

		select {
		case processedCh <- msg:
		case <-ctx.Done():
			return
		}
	}
}
