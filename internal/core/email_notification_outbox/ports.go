package email_notification_outbox

import (
	"context"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/google/uuid"
)

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}

// storer declares the behavior this package needs to persist and retrieve data.
type storer interface {
	CreateNotificationOutbox(
		ctx context.Context,
		arg db.CreateNotificationOutboxParams,
	) (db.EmailNotificationOutbox, error)
	GetUnprocessedNotificationOutbox(
		ctx context.Context,
		arg db.GetUnprocessedNotificationOutboxParams,
	) ([]db.EmailNotificationOutbox, error)
	MarkProcessedNotificationOutbox(
		ctx context.Context,
		arg db.MarkProcessedNotificationOutboxParams,
	) error
	IncrementRetryNotificationOutbox(
		ctx context.Context,
		ids []uuid.UUID,
	) error
}
