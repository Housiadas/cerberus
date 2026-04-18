// Package indexer defines stream index handler interfaces and implementations
// for consuming Kafka topics and indexing data into Elasticsearch.
package indexer

import (
	"context"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// StreamIndexHandler defines the contract for a Kafka-to-Elasticsearch indexing handler.
// Each implementation handles a single topic and its corresponding ES index.
type StreamIndexHandler interface {
	// Topic returns the Kafka topic this handler consumes.
	Topic() string
	// IndexName returns the Elasticsearch index name.
	IndexName() string
	// IndexMapping returns the JSON mapping for the Elasticsearch index.
	IndexMapping() string
	// Handle validates a single Kafka message before batching.
	Handle(ctx context.Context, msg *ckafka.Message) error
	// Flush processes a batch of validated messages and indexes them into Elasticsearch.
	Flush(ctx context.Context, msgs []*ckafka.Message) error
}
