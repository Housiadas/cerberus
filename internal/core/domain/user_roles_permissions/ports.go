package user_roles_permissions

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to retrieve data from the view.
type Storer interface {
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page web.Page) ([]UserRolesPermissions, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	HasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error)
}
