package outbox

import (
	"context"
	"time"

	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, o Outbox) error
	QueryUnprocessed(ctx context.Context, limit int, maxRetries int) ([]Outbox, error)
	MarkProcessed(ctx context.Context, ids []uuid.UUID, processedAt time.Time) error
	IncrementRetryCount(ctx context.Context, ids []uuid.UUID) error
}
