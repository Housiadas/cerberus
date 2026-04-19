package outbox

import (
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toCreateOutboxParams(
	id uuid.UUID,
	no NewOutbox,
	payload []byte,
	now time.Time,
) db.CreateOutboxParams {
	return db.CreateOutboxParams{
		ID:          id,
		EventType:   no.EventType,
		AggregateID: no.AggregateID,
		Topic:       no.Topic,
		Payload:     payload,
		CreatedAt:   pgtype.Timestamp{Time: now, Valid: true},
	}
}

func toDomainOutbox(o db.Outbox) Outbox {
	var processedAt *time.Time
	if o.ProcessedAt.Valid {
		processedAt = &o.ProcessedAt.Time
	}

	return New(
		o.ID,
		o.EventType,
		o.AggregateID,
		o.Topic,
		o.Payload,
		int(o.RetryCount),
		o.CreatedAt.Time,
		processedAt,
	)
}
