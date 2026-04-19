package indexer

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/sdk/elasticsearch"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/pkg/kafka"
	"github.com/Housiadas/cerberus/pkg/logger"
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// UserIndexHandler handles user events from Kafka and indexes them into Elasticsearch.
type UserIndexHandler struct {
	log logger.Logger
	es  *elasticsearch.Client
}

// NewUserHandler creates a new UserIndexHandler.
func NewUserHandler(log logger.Logger, es *elasticsearch.Client) *UserIndexHandler {
	return &UserIndexHandler{
		log: log,
		es:  es,
	}
}

func (h *UserIndexHandler) Topic() string        { return event.UserTopic }
func (h *UserIndexHandler) IndexName() string    { return elasticsearch.UserIndex }
func (h *UserIndexHandler) IndexMapping() string { return elasticsearch.UserMapping }

// Handle validates a single Kafka message. Only known event types pass through
// to the batcher; unknown types are rejected.
func (h *UserIndexHandler) Handle(_ context.Context, msg *ckafka.Message) error {
	eventType := kafka.ExtractHeader(msg, "event-type")

	switch eventType {
	case event.UserCreated, event.UserUpdated, event.UserDeleted:
		return nil
	default:
		return fmt.Errorf("unknown event type: %s", eventType)
	}
}

// Flush processes a batch of validated Kafka messages and indexes them into
// Elasticsearch in a single bulk request.
func (h *UserIndexHandler) Flush(ctx context.Context, msgs []*ckafka.Message) error {
	if len(msgs) == 0 {
		return nil
	}

	docs := make([]elasticsearch.Document, 0, len(msgs))

	for _, msg := range msgs {
		eventType := kafka.ExtractHeader(msg, "event-type")
		docID := string(msg.Key)

		switch eventType {
		case event.UserCreated, event.UserUpdated:
			docs = append(docs, elasticsearch.Document{
				Action: "index",
				ID:     docID,
				Body:   msg.Value,
			})
		case event.UserDeleted:
			docs = append(docs, elasticsearch.Document{
				Action: "delete",
				ID:     docID,
			})
		}
	}

	err := h.es.BulkIndex(ctx, h.IndexName(), docs)
	if err != nil {
		return fmt.Errorf("bulk indexing users: %w", err)
	}

	return nil
}
