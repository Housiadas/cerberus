package refund

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, ref Refund) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]Refund, error)
	QueryByID(ctx context.Context, refundID uuid.UUID) (Refund, error)
	QueryByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]Refund, error)
}
