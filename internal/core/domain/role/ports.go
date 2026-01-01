package role

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, role Role) error
	Update(ctx context.Context, role Role) error
	Delete(ctx context.Context, role Role) error
	Count(ctx context.Context, filter QueryFilter) (int, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page web.Page) ([]Role, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (Role, error)
}
