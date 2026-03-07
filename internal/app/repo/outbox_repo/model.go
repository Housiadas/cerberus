package outbox_repo

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/outbox"
	"github.com/google/uuid"
)

type outboxDB struct {
	ID          uuid.UUID       `db:"id"`
	EventType   string          `db:"event_type"`
	AggregateID uuid.UUID       `db:"aggregate_id"`
	Topic       string          `db:"topic"`
	Payload     json.RawMessage `db:"payload"`
	RetryCount  int             `db:"retry_count"`
	CreatedAt   time.Time       `db:"created_at"`
	ProcessedAt sql.NullTime    `db:"processed_at"`
}

func toOutboxDB(o outbox.Outbox) outboxDB {
	return outboxDB{
		ID:          o.ID(),
		EventType:   o.EventType(),
		AggregateID: o.AggregateID(),
		Topic:       o.Topic(),
		Payload:     o.Payload(),
		RetryCount:  o.RetryCount(),
		CreatedAt:   o.CreatedAt().UTC(),
		ProcessedAt: toNullTime(o.ProcessedAt()),
	}
}

func toOutboxDomain(db outboxDB) outbox.Outbox {
	return outbox.New(
		db.ID,
		db.EventType,
		db.AggregateID,
		db.Topic,
		db.Payload,
		db.RetryCount,
		db.CreatedAt.In(time.UTC),
		fromNullTime(db.ProcessedAt),
	)
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: t.UTC(), Valid: true}
}

func fromNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}

	t := nt.Time.In(time.UTC)

	return &t
}
