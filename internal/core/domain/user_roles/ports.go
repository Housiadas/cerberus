package user_roles

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	Create(ctx context.Context, ur UserRole) error
	Delete(ctx context.Context, ur UserRole) error
	Count(ctx context.Context, filter QueryFilter) (int, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]UserRole, error)
	GetUserRoleNames(ctx context.Context, userID uuid.UUID) ([]string, error)
}
