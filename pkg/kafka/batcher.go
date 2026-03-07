package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type offsetStorer interface {
	StoreMessage(m *kafka.Message) ([]kafka.TopicPartition, error)
}

// Batcher accumulates processed messages and flushes them in bulk.
type Batcher struct {
	storer       offsetStorer
	flusher      Flusher
	batchSize    int
	flushTimeout time.Duration
	log          *logger.Service
}

// NewBatcher creates a new Batcher.
func NewBatcher(
	storer offsetStorer,
	flusher Flusher,
	batchSize int,
	timeout time.Duration,
	log *logger.Service,
) *Batcher {
	return &Batcher{
		storer:       storer,
		flusher:      flusher,
		batchSize:    batchSize,
		flushTimeout: timeout,
		log:          log,
	}
}

// Run reads from processedCh and flushes batches. It returns when processedCh is closed.
func (b *Batcher) Run(ctx context.Context, processedCh <-chan *kafka.Message) error {
	ticker := time.NewTicker(b.flushTimeout)
	defer ticker.Stop()

	batch := make([]*kafka.Message, 0, b.batchSize)

	for {
		select {
		case msg, ok := <-processedCh:
			if !ok {
				// channel closed: flush remaining
				if len(batch) > 0 {
					return b.flush(ctx, batch)
				}
				return nil
			}

			batch = append(batch, msg)
			if len(batch) >= b.batchSize {
				if err := b.flush(ctx, batch); err != nil {
					return err
				}
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				if err := b.flush(ctx, batch); err != nil {
					return err
				}
				batch = batch[:0]
			}
		}
	}
}

func (b *Batcher) flush(ctx context.Context, batch []*kafka.Message) error {
	if err := b.flusher(ctx, batch); err != nil {
		return fmt.Errorf("batcher: flush error: %w", err)
	}

	for _, msg := range batch {
		if _, storeErr := b.storer.StoreMessage(msg); storeErr != nil {
			b.log.Error(ctx, fmt.Sprintf("batcher: store offset error: %v", storeErr))
		}
	}

	return nil
}
