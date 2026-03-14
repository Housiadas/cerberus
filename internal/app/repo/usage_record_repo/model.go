package usage_record_repo

import (
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/usage_record"
	"github.com/google/uuid"
)

type usageRecordDB struct {
	ID             uuid.UUID `db:"id"`
	SubscriptionID uuid.UUID `db:"subscription_id"`
	Description    string    `db:"description"`
	Quantity       int       `db:"quantity"`
	RecordedAt     time.Time `db:"recorded_at"`
	CreatedAt      time.Time `db:"created_at"`
}

func toUsageRecordDB(ur usage_record.UsageRecord) usageRecordDB {
	return usageRecordDB{
		ID:             ur.ID(),
		SubscriptionID: ur.SubscriptionID(),
		Description:    ur.Description(),
		Quantity:       ur.Quantity(),
		RecordedAt:     ur.RecordedAt().UTC(),
		CreatedAt:      ur.CreatedAt().UTC(),
	}
}

func toUsageRecordDomain(db usageRecordDB) usage_record.UsageRecord {
	return usage_record.New(
		db.ID,
		db.SubscriptionID,
		db.Description,
		db.Quantity,
		db.RecordedAt.In(time.UTC),
		db.CreatedAt.In(time.UTC),
	)
}

func toUsageRecordsDomain(dbs []usageRecordDB) []usage_record.UsageRecord {
	urs := make([]usage_record.UsageRecord, len(dbs))

	for i, db := range dbs {
		urs[i] = toUsageRecordDomain(db)
	}

	return urs
}
