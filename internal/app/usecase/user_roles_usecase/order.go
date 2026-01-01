package user_roles_usecase

import (
	ur "github.com/Housiadas/cerberus/internal/core/domain/user_roles"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	"user_id": ur.OrderByUserID,
	"role_id": ur.OrderByRoleID,
}

var defaultOrderBy = order.By{
	Field:     ur.OrderByUserID,
	Direction: order.ASC,
}
