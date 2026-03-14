package payment

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, pmt Payment) error
	Update(ctx context.Context, pmt Payment) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]Payment, error)
	QueryByID(ctx context.Context, paymentID uuid.UUID) (Payment, error)
	QueryByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]Payment, error)
}
