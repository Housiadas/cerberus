package user_roles_permissions

import (
	"context"

	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to retrieve data from the view.
type Storer interface {
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		page page.Page,
	) ([]UserRolesPermissions, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	HasPermission(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error)
	QueryPermissionsByUserID(ctx context.Context, userID uuid.UUID) ([]Permission, error)
}
