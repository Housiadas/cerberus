package account_repo

import (
	"github.com/Housiadas/cerberus/internal/core/domain/account"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	account.OrderByID:   "id",
	account.OrderByName: "name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", ErrFieldNotExist
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
