package kafka_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgkafka "github.com/Housiadas/cerberus/pkg/kafka"
	"github.com/Housiadas/cerberus/pkg/logger"
)

// fakeStorer is a no-op offsetStorer for tests.
type fakeStorer struct{}

func (f *fakeStorer) StoreMessage(_ *kafka.Message) ([]kafka.TopicPartition, error) {
	return nil, nil
}

func newTestLogger() *logger.Service {
	return logger.New(io.Discard, logger.LevelInfo, "TEST", nil, nil)
}

func makeMessages(n int) []*kafka.Message {
	msgs := make([]*kafka.Message, n)
	for i := range msgs {
		msgs[i] = &kafka.Message{}
	}
	return msgs
}

func feedChannel(ch chan<- *kafka.Message, msgs []*kafka.Message) {
	for _, m := range msgs {
		ch <- m
	}
}

func TestBatcher_FlushBySize(t *testing.T) {
	t.Parallel()

	const batchSize = 5
	flushed := make([][]*kafka.Message, 0)

	flusher := func(_ context.Context, msgs []*kafka.Message) error {
		cp := make([]*kafka.Message, len(msgs))
		copy(cp, msgs)
		flushed = append(flushed, cp)
		return nil
	}

	log := newTestLogger()
	b := pkgkafka.NewBatcher(&fakeStorer{}, flusher, batchSize, 10*time.Second, log)

	ch := make(chan *kafka.Message, batchSize)
	msgs := makeMessages(batchSize)
	feedChannel(ch, msgs)
	close(ch)

	err := b.Run(context.Background(), ch)
	require.NoError(t, err)
	assert.Len(t, flushed, 1)
	assert.Len(t, flushed[0], batchSize)
}

func TestBatcher_FlushByTimeout(t *testing.T) {
	t.Parallel()

	flushed := make([][]*kafka.Message, 0)
	flusher := func(_ context.Context, msgs []*kafka.Message) error {
		cp := make([]*kafka.Message, len(msgs))
		copy(cp, msgs)
		flushed = append(flushed, cp)
		return nil
	}

	log := newTestLogger()
	b := pkgkafka.NewBatcher(&fakeStorer{}, flusher, 100, 50*time.Millisecond, log)

	ch := make(chan *kafka.Message, 10)
	msgs := makeMessages(3)
	feedChannel(ch, msgs)

	// give the ticker time to fire, then close channel
	go func() {
		time.Sleep(200 * time.Millisecond)
		close(ch)
	}()

	err := b.Run(context.Background(), ch)
	require.NoError(t, err)
	// at least one flush triggered by ticker
	assert.NotEmpty(t, flushed)
	total := 0
	for _, batch := range flushed {
		total += len(batch)
	}
	assert.Equal(t, 3, total)
}

func TestBatcher_FlushOnClose(t *testing.T) {
	t.Parallel()

	flushed := make([][]*kafka.Message, 0)
	flusher := func(_ context.Context, msgs []*kafka.Message) error {
		cp := make([]*kafka.Message, len(msgs))
		copy(cp, msgs)
		flushed = append(flushed, cp)
		return nil
	}

	log := newTestLogger()
	b := pkgkafka.NewBatcher(&fakeStorer{}, flusher, 100, 10*time.Second, log)

	ch := make(chan *kafka.Message, 10)
	msgs := makeMessages(2)
	feedChannel(ch, msgs)
	close(ch)

	err := b.Run(context.Background(), ch)
	require.NoError(t, err)
	assert.Len(t, flushed, 1)
	assert.Len(t, flushed[0], 2)
}

func TestBatcher_FlusherError(t *testing.T) {
	t.Parallel()

	flusherErr := errors.New("db error")
	flusher := func(_ context.Context, _ []*kafka.Message) error {
		return flusherErr
	}

	log := newTestLogger()
	b := pkgkafka.NewBatcher(&fakeStorer{}, flusher, 1, 10*time.Second, log)

	ch := make(chan *kafka.Message, 2)
	ch <- &kafka.Message{}
	close(ch)

	err := b.Run(context.Background(), ch)
	require.Error(t, err)
	assert.ErrorIs(t, err, flusherErr)
}
