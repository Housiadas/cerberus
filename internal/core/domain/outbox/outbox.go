// Package outbox defines the outbox domain model.
package outbox

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Outbox represents an outbox event entry.
type Outbox struct {
	ID          uuid.UUID
	EventType   string
	AggregateID uuid.UUID
	Topic       string
	Payload     json.RawMessage
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

// NewOutbox contains information needed to create a new outbox entry.
type NewOutbox struct {
	EventType   string
	AggregateID uuid.UUID
	Topic       string
	Payload     any
}
