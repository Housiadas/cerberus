package user_roles_permissions

import "github.com/Housiadas/cerberus/pkg/order"

// Set of fields that the results can be ordered by.
const (
	OrderByUserName       = "user_name"
	OrderByUserEmail      = "user_email"
	OrderByRoleName       = "role_name"
	OrderByPermissionName = "permission_name"
)

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByUserName, order.ASC)
