package email_notification_outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}

// Storer declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	Create(ctx context.Context, e EmailNotificationOutbox) error
	QueryUnprocessed(
		ctx context.Context,
		limit int,
		maxRetries int,
	) ([]EmailNotificationOutbox, error)
	MarkProcessed(ctx context.Context, ids []uuid.UUID, processedAt time.Time) error
	IncrementRetryCount(ctx context.Context, ids []uuid.UUID) error
}
