package billing_address_repo

import (
	"github.com/Housiadas/cerberus/internal/core/domain/billing_address"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	billing_address.OrderByID: "id",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", ErrFieldNotExist
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
