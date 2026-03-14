package payment_method

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, pm PaymentMethod) error
	Delete(ctx context.Context, pm PaymentMethod) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]PaymentMethod, error)
	QueryByID(ctx context.Context, paymentMethodID uuid.UUID) (PaymentMethod, error)
	QueryByAccountID(ctx context.Context, accountID uuid.UUID) ([]PaymentMethod, error)
}
