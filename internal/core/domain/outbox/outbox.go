// Package outbox defines the outbox domain model.
package outbox

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Outbox represents an outbox event entry.
type Outbox struct {
	id          uuid.UUID
	eventType   string
	aggregateID uuid.UUID
	topic       string
	payload     json.RawMessage
	retryCount  int
	createdAt   time.Time
	processedAt *time.Time
}

// New constructs an Outbox from its individual fields.
func New(
	id uuid.UUID,
	eventType string,
	aggregateID uuid.UUID,
	topic string,
	payload json.RawMessage,
	retryCount int,
	createdAt time.Time,
	processedAt *time.Time,
) Outbox {
	return Outbox{
		id:          id,
		eventType:   eventType,
		aggregateID: aggregateID,
		topic:       topic,
		payload:     payload,
		retryCount:  retryCount,
		createdAt:   createdAt,
		processedAt: processedAt,
	}
}

// ID returns the outbox entry ID.
func (o Outbox) ID() uuid.UUID { return o.id }

// EventType returns the event type.
func (o Outbox) EventType() string { return o.eventType }

// AggregateID returns the aggregate ID.
func (o Outbox) AggregateID() uuid.UUID { return o.aggregateID }

// Topic returns the Kafka topic.
func (o Outbox) Topic() string { return o.topic }

// Payload returns the message payload.
func (o Outbox) Payload() json.RawMessage { return o.payload }

// RetryCount returns the retry count.
func (o Outbox) RetryCount() int { return o.retryCount }

// CreatedAt returns the creation time.
func (o Outbox) CreatedAt() time.Time { return o.createdAt }

// ProcessedAt returns the processing time (nil if not yet processed).
func (o Outbox) ProcessedAt() *time.Time { return o.processedAt }

// NewOutbox contains information needed to create a new outbox entry.
type NewOutbox struct {
	EventType   string
	AggregateID uuid.UUID
	Topic       string
	Payload     any
}
