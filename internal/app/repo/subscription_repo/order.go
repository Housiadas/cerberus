package subscription_repo

import (
	"github.com/Housiadas/cerberus/internal/core/domain/subscription"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	subscription.OrderByID:        "id",
	subscription.OrderByCreatedAt: "created_at",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", ErrFieldNotExist
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
