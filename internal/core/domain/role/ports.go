package role

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, role Role) error
	Update(ctx context.Context, role Role) error
	Delete(ctx context.Context, role Role) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Role, error)
}
