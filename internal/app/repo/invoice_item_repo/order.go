package invoice_item_repo

import (
	"github.com/Housiadas/cerberus/internal/core/domain/invoice_item"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	invoice_item.OrderByID:        "id",
	invoice_item.OrderByCreatedAt: "created_at",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", ErrFieldNotExist
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
