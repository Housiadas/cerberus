package role_repo

import (
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/pkg/order"
)

//nolint:gochecknoglobals
var orderByFields = map[string]string{
	role.OrderByName: "name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", ErrFieldNotExist
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
