package payment_repo

import (
	"github.com/Housiadas/cerberus/internal/core/domain/payment"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	payment.OrderByID:        "id",
	payment.OrderByCreatedAt: "created_at",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", ErrFieldNotExist
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
