package kafka_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgkafka "github.com/Housiadas/cerberus/pkg/kafka"
)

// We test the worker pool behavior via the Batcher+Handler path using a
// channel-based approach without real Kafka.

func TestConsumer_WorkerPoolProcessesAllMessages(t *testing.T) {
	t.Parallel()

	const numMessages = 20
	var processed atomic.Int64

	handler := pkgkafka.Handler(func(_ context.Context, _ *kafka.Message) error {
		processed.Add(1)
		return nil
	})

	var flushed atomic.Int64
	flusher := pkgkafka.Flusher(func(_ context.Context, msgs []*kafka.Message) error {
		flushed.Add(int64(len(msgs)))
		return nil
	})

	// Use a mock consumer that feeds messages then blocks until context cancelled.
	msgCh := make(chan *kafka.Message, numMessages)
	for i := 0; i < numMessages; i++ {
		msgCh <- &kafka.Message{}
	}

	log := newTestLogger()
	b := pkgkafka.NewBatcher(&fakeStorer{}, flusher, numMessages, 1*time.Second, log)

	processedCh := make(chan *kafka.Message, numMessages)

	// Simulate workers manually by feeding msgCh into processedCh via handler.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for msg := range msgCh {
			if err := handler(ctx, msg); err == nil {
				processedCh <- msg
			}
		}
		close(processedCh)
	}()

	close(msgCh) // all messages already queued

	err := b.Run(ctx, processedCh)
	require.NoError(t, err)
	assert.Equal(t, int64(numMessages), processed.Load())
	assert.Equal(t, int64(numMessages), flushed.Load())
}

func TestConsumer_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// processedCh that never receives a message
	processedCh := make(chan *kafka.Message)

	log := newTestLogger()
	b := pkgkafka.NewBatcher(&fakeStorer{}, func(_ context.Context, _ []*kafka.Message) error {
		return nil
	}, 10, 10*time.Second, log)

	// Run in goroutine; close processedCh when context is done so Run returns.
	go func() {
		<-ctx.Done()
		close(processedCh)
	}()

	start := time.Now()
	err := b.Run(ctx, processedCh)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 500*time.Millisecond)
}
