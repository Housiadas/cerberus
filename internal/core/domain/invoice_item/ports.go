package invoice_item

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, item InvoiceItem) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]InvoiceItem, error)
	QueryByID(ctx context.Context, invoiceItemID uuid.UUID) (InvoiceItem, error)
	QueryByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]InvoiceItem, error)
}
