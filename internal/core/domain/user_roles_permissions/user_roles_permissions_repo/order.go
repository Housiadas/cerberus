package user_roles_permissions_repo

import (
	"fmt"

	urp "github.com/Housiadas/cerberus/internal/core/domain/user_roles_permissions"
	"github.com/Housiadas/cerberus/pkg/order"
)

func orderByClause(ob order.By) (string, error) {
	by, exists := getOrderFields()[ob.Field]
	if !exists {
		return "", fmt.Errorf("%w: %s", errOrderFieldNotFound, ob.Field)
	}

	return " ORDER BY " + by + " " + ob.Direction, nil
}

func getOrderFields() map[string]string {
	return map[string]string{
		urp.OrderByUserName:       "user_name",
		urp.OrderByUserEmail:      "user_email",
		urp.OrderByRoleName:       "role_name",
		urp.OrderByPermissionName: "permission_name",
	}
}
