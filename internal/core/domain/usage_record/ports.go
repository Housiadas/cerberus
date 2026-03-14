package usage_record

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, ur UsageRecord) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]UsageRecord, error)
	QueryByID(ctx context.Context, usageRecordID uuid.UUID) (UsageRecord, error)
}
