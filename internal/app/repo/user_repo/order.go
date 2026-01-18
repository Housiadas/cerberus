package user_repo

import (
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/pkg/order"
)

//nolint:gochecknoglobals
var orderByFields = map[string]string{
	user.OrderByID:      "id",
	user.OrderByName:    "name",
	user.OrderByEmail:   "email",
	user.OrderByEnabled: "enabled",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("%w: %s", errOrderFieldNotFound, orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
