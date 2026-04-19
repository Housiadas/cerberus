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

// ConsumerConfig holds configuration for Consumer.
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

// Consumer is a Kafka consumer with a worker pool and batcher.
type Consumer struct {
	consumer     *kafka.Consumer
	log          logger.Logger
	workers      int
	batchSize    int
	flushTimeout time.Duration
	bufferSize   int
}

// NewConsumer creates a new Consumer.
func NewConsumer(log logger.Logger, cfg ConsumerConfig) (*Consumer, error) {
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
		"security.protocol":        cfg.SecurityProtocol,
		"session.timeout.ms":       cfg.SessionTimeout,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       true,
		"enable.auto.offset.store": false, // offsets are stored manually after batch flush
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &Consumer{
		consumer:     consumer,
		log:          log,
		workers:      cfg.Workers,
		batchSize:    cfg.BatchSize,
		flushTimeout: cfg.FlushTimeout,
		bufferSize:   cfg.BufferSize,
	}, nil
}

// Close shuts down the underlying Kafka consumer.
func (c *Consumer) Close() {
	c.consumer.Close()
}

// Subscribe subscribes to the given Kafka topic.
func (c *Consumer) Subscribe(topic string) error {
	err := c.consumer.Subscribe(topic, nil)
	if err != nil {
		return fmt.Errorf("kafka subscribe error: %w", err)
	}

	return nil
}

// SubscribeTopics subscribes to the given Kafka topics
func (c *Consumer) SubscribeTopics(topics []string) error {
	err := c.consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return fmt.Errorf("kafka subscribe topics error: %w", err)
	}

	return nil
}

// Consume starts the pipeline: poller → workers → batcher.
// It blocks until the context is canceled or a fatal error occurs.
//
// Shutdown order: poller stops → msgCh closes → workers drain and exit →
// processedCh closes → batcher flushes remaining and exits.
func (c *Consumer) Consume(
	ctx context.Context,
	handler Handler,
	flusher Flusher,
) error {
	// Pipeline channels:
	// msgCh: raw messages from Kafka poller → worker pool
	// processedCh: successfully handled messages from workers → batcher
	msgCh := make(chan *kafka.Message, c.bufferSize)
	processedCh := make(chan *kafka.Message, c.bufferSize)

	// Stage 1: Worker pool each worker reads from msgCh, calls handler,
	// and forwards successfully processed messages to processedCh.
	// Failed messages are logged and skipped (not forwarded), so their
	// offsets won't be committed, and they will be re-delivered.
	var wg sync.WaitGroup
	for range c.workers {
		wg.Go(func() {
			c.runWorker(ctx, msgCh, processedCh, handler)
		})
	}

	// Stage 2: Batcher — collects processed messages and flushes them in
	// bulk via flusher. After a successful flush, offsets are stored so
	// auto-commit can pick them up (at-least-once delivery guarantee).
	batcher := NewBatcher(c.consumer, flusher, c.batchSize, c.flushTimeout, c.log)

	batchErrCh := make(chan error, 1)

	go func() {
		batchErrCh <- batcher.Run(ctx, processedCh)
	}()

	// Stage 0: Poller — polls Kafka for events and dispatches messages to msgCh.
	// Blocks until ctx is canceled or a fatal Kafka error occurs.
	pollErr := c.runPoller(ctx, msgCh)

	// Graceful shutdown in pipeline order:
	// 1. close(msgCh)    — signals workers there are no more messages
	// 2. wg.Wait()       — waits for workers to finish in-flight handling
	// 3. close(processedCh) — signals batcher to flush remaining and exit
	close(msgCh)
	wg.Wait()
	close(processedCh)

	// Wait for batcher to finish its final flush.
	batchErr := <-batchErrCh

	if pollErr != nil {
		return pollErr
	}

	return batchErr
}

// runPoller continuously polls Kafka for events.
// Poll(200) blocks up to 100ms per call; the select on ctx.Done()
// between polls ensures we react to cancellation promptly.
func (c *Consumer) runPoller(
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

// handleEvent dispatches a single Kafka event.
// Messages are sent to msgCh for worker processing.
// Fatal errors stop the consumer; retriable errors are logged and skipped.
func (c *Consumer) handleEvent(
	ctx context.Context,
	ev kafka.Event,
	msgCh chan<- *kafka.Message,
) error {
	switch event := ev.(type) {
	case nil:
		// Poll timeout expired with no event — this is normal.
	case *kafka.Message:
		// Forward a message to the worker pool. If the context is canceled
		// while msgCh is full, we exit instead of blocking forever.
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

// runWorker processes messages from msgCh one at a time.
// Successfully handled messages are forwarded to processedCh for batching.
// On handler error the message is skipped — its offset won't be stored,
// so Kafka will re-deliver it on the next consumer startup.
func (c *Consumer) runWorker(
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
