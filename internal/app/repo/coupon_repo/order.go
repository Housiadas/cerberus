package coupon_repo

import (
	"github.com/Housiadas/cerberus/internal/core/domain/coupon"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	coupon.OrderByID:        "id",
	coupon.OrderByCode:      "code",
	coupon.OrderByCreatedAt: "created_at",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", ErrFieldNotExist
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
