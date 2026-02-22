package user_roles_permissions_usecase

import (
	urp "github.com/Housiadas/cerberus/internal/core/domain/user_roles_permissions"
	"github.com/Housiadas/cerberus/pkg/order"
)

func getDefaultOrderBy() order.By {
	return order.By{
		Field:     urp.OrderByUserName,
		Direction: order.ASC,
	}
}

func getOrderByFields() map[string]string {
	return map[string]string{
		"user_name":       urp.OrderByUserName,
		"user_email":      urp.OrderByUserEmail,
		"role_name":       urp.OrderByRoleName,
		"permission_name": urp.OrderByPermissionName,
	}
}
