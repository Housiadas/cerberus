package invoice

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, inv Invoice) error
	Update(ctx context.Context, inv Invoice) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]Invoice, error)
	QueryByID(ctx context.Context, invoiceID uuid.UUID) (Invoice, error)
	QueryByAccountID(ctx context.Context, accountID uuid.UUID) ([]Invoice, error)
}
