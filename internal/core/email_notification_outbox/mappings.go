package email_notification_outbox

import (
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toCreateNotificationOutboxParams(
	id uuid.UUID,
	no NewEmailNotificationOutbox,
	payload []byte,
	now time.Time,
) db.CreateNotificationOutboxParams {
	return db.CreateNotificationOutboxParams{
		ID:        id,
		EventType: no.EventType,
		ToEmail:   no.ToEmail,
		Payload:   payload,
		CreatedAt: pgtype.Timestamp{Time: now, Valid: true},
	}
}

func toDomainEmailNotificationOutbox(o db.EmailNotificationOutbox) EmailNotificationOutbox {
	var processedAt *time.Time
	if o.ProcessedAt.Valid {
		processedAt = &o.ProcessedAt.Time
	}

	return New(
		o.ID,
		o.EventType,
		o.ToEmail,
		o.Payload,
		int(o.RetryCount),
		o.CreatedAt.Time,
		processedAt,
	)
}

func toDomainEmailNotificationOutboxes(dbEntries []db.EmailNotificationOutbox) []EmailNotificationOutbox {
	entries := make([]EmailNotificationOutbox, len(dbEntries))
	for i, o := range dbEntries {
		entries[i] = toDomainEmailNotificationOutbox(o)
	}

	return entries
}
