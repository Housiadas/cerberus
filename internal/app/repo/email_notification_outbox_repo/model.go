package email_notification_outbox_repo

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/email_notification_outbox"
	"github.com/google/uuid"
)

type emailOutboxDB struct {
	ID          uuid.UUID       `db:"id"`
	EventType   string          `db:"event_type"`
	ToEmail     string          `db:"to_email"`
	Payload     json.RawMessage `db:"payload"`
	RetryCount  int             `db:"retry_count"`
	CreatedAt   time.Time       `db:"created_at"`
	ProcessedAt sql.NullTime    `db:"processed_at"`
}

func toEmailOutboxDB(e email_notification_outbox.EmailNotificationOutbox) emailOutboxDB {
	return emailOutboxDB{
		ID:          e.ID(),
		EventType:   e.EventType(),
		ToEmail:     e.ToEmail(),
		Payload:     e.Payload(),
		RetryCount:  e.RetryCount(),
		CreatedAt:   e.CreatedAt().UTC(),
		ProcessedAt: toNullTime(e.ProcessedAt()),
	}
}

func toEmailOutboxDomain(db emailOutboxDB) email_notification_outbox.EmailNotificationOutbox {
	return email_notification_outbox.New(
		db.ID,
		db.EventType,
		db.ToEmail,
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
