// Package relay implements the outbox relay that polls the outbox table
// and produces messages to Kafka.
package relay

import (
	"context"
	"time"

	"github.com/Housiadas/cerberus/internal/core/outbox/outbox_service"
	"github.com/Housiadas/cerberus/pkg/kafka"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

const (
	flushTimeoutMs = 10_000
)

// Relay polls the outbox table and produces messages to Kafka.
type Relay struct {
	log        logger.Logger
	outboxSvc  *outbox_service.Service
	producer   kafka.Producer
	interval   time.Duration
	batchSize  int
	maxRetries int
}

// New constructs a new Relay.
func New(
	log logger.Logger,
	outboxSvc *outbox_service.Service,
	producer kafka.Producer,
	interval time.Duration,
	batchSize int,
	maxRetries int,
) *Relay {
	return &Relay{
		log:        log,
		outboxSvc:  outboxSvc,
		producer:   producer,
		interval:   interval,
		batchSize:  batchSize,
		maxRetries: maxRetries,
	}
}

// Start begins the relay polling loop. It blocks until the context is canceled.
func (r *Relay) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.log.Info(
		ctx,
		"relay",
		"status",
		"started",
		"interval",
		r.interval.String(),
		"batchSize",
		r.batchSize,
		"maxRetries",
		r.maxRetries,
	)

	for {
		select {
		case <-ctx.Done():
			r.log.Info(ctx, "relay", "status", "shutting down, flushing producer")
			r.producer.Flush(flushTimeoutMs)
			r.log.Info(ctx, "relay", "status", "stopped")

			return
		case <-ticker.C:
			r.processBatch(ctx)
		}
	}
}

func (r *Relay) processBatch(ctx context.Context) {
	entries, err := r.outboxSvc.QueryUnprocessed(ctx, r.batchSize, r.maxRetries)
	if err != nil {
		r.log.Error(ctx, "relay", "status", "query unprocessed error", "msg", err)

		return
	}

	if len(entries) == 0 {
		return
	}

	var (
		failedIDs    []uuid.UUID
		processedIDs []uuid.UUID
	)

	for _, entry := range entries {
		msg := &ckafka.Message{
			TopicPartition: ckafka.TopicPartition{
				Topic:     new(entry.Topic()),
				Partition: ckafka.PartitionAny,
			},
			Key:   []byte(entry.AggregateID().String()),
			Value: entry.Payload(),
		}

		err := r.producer.Produce(ctx, msg)
		if err != nil {
			r.log.Error(
				ctx,
				"relay",
				"status", "produce error",
				"event_id", entry.ID(),
				"msg", err,
			)
			failedIDs = append(failedIDs, entry.ID())

			continue
		}

		processedIDs = append(processedIDs, entry.ID())
	}

	telemetry.AddSpan(ctx,
		"outbox.processedIDs",
		attribute.Int("processedIDs", len(processedIDs)),
	)
	r.producer.Flush(flushTimeoutMs)

	r.log.Info(
		ctx,
		"relay",
		"status", "batch processed",
		"total", len(entries),
		"produced", len(processedIDs),
		"failed", len(failedIDs),
	)

	if len(failedIDs) > 0 {
		telemetry.AddSpan(ctx, "outbox.failedIDs", attribute.Int("failedIDs", len(failedIDs)))

		err := r.outboxSvc.IncrementRetryCount(ctx, failedIDs)
		if err != nil {
			r.log.Error(
				ctx,
				"relay",
				"status", "increment retry count error",
				"msg", err,
			)
		}
	}

	if len(processedIDs) == 0 {
		return
	}

	err = r.outboxSvc.MarkProcessed(ctx, processedIDs)
	if err != nil {
		r.log.Error(ctx,
			"relay",
			"status", "mark processed error",
			"msg", err,
		)
	}
}
