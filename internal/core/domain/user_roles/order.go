package user_roles

import "github.com/Housiadas/cerberus/pkg/order"

// Set of fields that the results can be ordered by.
const (
	OrderByUserID = "user_id"
	OrderByRoleID = "role_id"
)

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByUserID, order.ASC)
