package coupon

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, cpn Coupon) error
	Update(ctx context.Context, cpn Coupon) error
	Delete(ctx context.Context, cpn Coupon) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]Coupon, error)
	QueryByID(ctx context.Context, couponID uuid.UUID) (Coupon, error)
	QueryByCode(ctx context.Context, code string) (Coupon, error)
}
