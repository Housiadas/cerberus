package tax_rate_repo

import (
	"github.com/Housiadas/cerberus/internal/core/domain/tax_rate"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	tax_rate.OrderByID:   "id",
	tax_rate.OrderByName: "name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", ErrFieldNotExist
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
