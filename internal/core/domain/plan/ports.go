package plan

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, pl Plan) error
	Update(ctx context.Context, pl Plan) error
	Delete(ctx context.Context, pl Plan) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]Plan, error)
	QueryByID(ctx context.Context, planID uuid.UUID) (Plan, error)
}
