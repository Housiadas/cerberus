// Package email_notification_outbox defines the email notification outbox domain model.
package email_notification_outbox

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EmailNotificationOutbox represents an email notification outbox entry.
type EmailNotificationOutbox struct {
	id          uuid.UUID
	eventType   string
	toEmail     string
	payload     json.RawMessage
	retryCount  int
	createdAt   time.Time
	processedAt *time.Time
}

// New constructs an EmailNotificationOutbox from its individual fields.
func New(
	id uuid.UUID,
	eventType string,
	toEmail string,
	payload json.RawMessage,
	retryCount int,
	createdAt time.Time,
	processedAt *time.Time,
) EmailNotificationOutbox {
	return EmailNotificationOutbox{
		id:          id,
		eventType:   eventType,
		toEmail:     toEmail,
		payload:     payload,
		retryCount:  retryCount,
		createdAt:   createdAt,
		processedAt: processedAt,
	}
}

// ID returns the entry ID.
func (e EmailNotificationOutbox) ID() uuid.UUID { return e.id }

// EventType returns the event type.
func (e EmailNotificationOutbox) EventType() string { return e.eventType }

// ToEmail returns the recipient email address.
func (e EmailNotificationOutbox) ToEmail() string { return e.toEmail }

// Payload returns the raw JSON payload.
func (e EmailNotificationOutbox) Payload() json.RawMessage { return e.payload }

// RetryCount returns the retry count.
func (e EmailNotificationOutbox) RetryCount() int { return e.retryCount }

// CreatedAt returns the creation time.
func (e EmailNotificationOutbox) CreatedAt() time.Time { return e.createdAt }

// ProcessedAt returns the processing time (nil if not yet processed).
func (e EmailNotificationOutbox) ProcessedAt() *time.Time { return e.processedAt }

// NewEmailNotificationOutbox contains information needed to create a new email notification entry.
type NewEmailNotificationOutbox struct {
	EventType string
	ToEmail   string
	Payload   any
}
