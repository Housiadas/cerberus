package payment_method_repo

import (
	"github.com/Housiadas/cerberus/internal/core/domain/payment_method"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	payment_method.OrderByID:        "id",
	payment_method.OrderByCreatedAt: "created_at",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", ErrFieldNotExist
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
