package user_roles_permissions

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to retrieve data from the view.
type Storer interface {
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]UserRolesPermissions, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	HasPermission(ctx context.Context, userID uuid.UUID, permissionID uuid.UUID) (bool, error)
}
