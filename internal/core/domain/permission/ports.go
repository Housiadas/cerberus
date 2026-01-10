package permission

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, p Permission) error
	Update(ctx context.Context, p Permission) error
	Delete(ctx context.Context, p Permission) error
	Count(ctx context.Context, filter QueryFilter) (int, error)
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		page web.Page,
	) ([]Permission, error)
	QueryByID(ctx context.Context, permissionID uuid.UUID) (Permission, error)
}
