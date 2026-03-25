package account_repo

import (
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	account.OrderByID:   "id",
	account.OrderByName: "name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("%w: %s", errOrderFieldNotFound, orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
