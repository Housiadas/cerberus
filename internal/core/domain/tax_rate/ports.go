package tax_rate

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, tr TaxRate) error
	Update(ctx context.Context, tr TaxRate) error
	Delete(ctx context.Context, tr TaxRate) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]TaxRate, error)
	QueryByID(ctx context.Context, taxRateID uuid.UUID) (TaxRate, error)
}
