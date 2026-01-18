package user_roles_permissions_repo

import (
	"fmt"

	urp "github.com/Housiadas/cerberus/internal/core/domain/user_roles_permissions"
	"github.com/Housiadas/cerberus/pkg/order"
)

//nolint:gochecknoglobals
var orderByFields = map[string]string{
	urp.OrderByUserName:       "user_name",
	urp.OrderByUserEmail:      "user_email",
	urp.OrderByRoleName:       "role_name",
	urp.OrderByPermissionName: "permission_name",
}

func orderByClause(ob order.By) (string, error) {
	by, exists := orderByFields[ob.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", ob.Field)
	}

	return " ORDER BY " + by + " " + ob.Direction, nil
}
