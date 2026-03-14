package usage_record

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("usage record not found")

type UsageRecord struct {
	id             uuid.UUID
	subscriptionID uuid.UUID
	description    string
	quantity       int
	recordedAt     time.Time
	createdAt      time.Time
}

func New(
	id uuid.UUID,
	subscriptionID uuid.UUID,
	description string,
	quantity int,
	recordedAt time.Time,
	createdAt time.Time,
) UsageRecord {
	return UsageRecord{
		id: id, subscriptionID: subscriptionID, description: description,
		quantity: quantity, recordedAt: recordedAt, createdAt: createdAt,
	}
}

func (u UsageRecord) ID() uuid.UUID             { return u.id }
func (u UsageRecord) SubscriptionID() uuid.UUID { return u.subscriptionID }
func (u UsageRecord) Description() string       { return u.description }
func (u UsageRecord) Quantity() int             { return u.quantity }
func (u UsageRecord) RecordedAt() time.Time     { return u.recordedAt }
func (u UsageRecord) CreatedAt() time.Time      { return u.createdAt }

type NewUsageRecord struct {
	SubscriptionID uuid.UUID
	Description    string
	Quantity       int
	RecordedAt     time.Time
}

type QueryFilter struct {
	ID             *uuid.UUID
	SubscriptionID *uuid.UUID
}
