package user_roles_repo

import (
	"fmt"

	ur "github.com/Housiadas/cerberus/internal/core/domain/user_roles"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	ur.OrderByUserID: "user_id",
	ur.OrderByRoleID: "role_id",
}

func orderByClause(ob order.By) (string, error) {
	by, exists := orderByFields[ob.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", ob.Field)
	}
	return " ORDER BY " + by + " " + ob.Direction, nil
}
