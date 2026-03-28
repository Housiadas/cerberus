package permission

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	Create(ctx context.Context, p Permission) error
	Update(ctx context.Context, p Permission) error
	Delete(ctx context.Context, p Permission) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]Permission, error)
	QueryByID(ctx context.Context, permissionID uuid.UUID) (Permission, error)
}
