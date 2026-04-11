package outbox

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

// storer interface declares the behavior this package needs to persist and retrieve data.
type storer interface {
	CreateOutbox(ctx context.Context, arg db.CreateOutboxParams) (db.Outbox, error)
	GetUnprocessedOutbox(ctx context.Context, arg db.GetUnprocessedOutboxParams) ([]db.Outbox, error)
	MarkProcessedOutbox(ctx context.Context, arg db.MarkProcessedOutboxParams) error
	IncrementRetryOutbox(ctx context.Context, ids uuid.UUID) error
}
