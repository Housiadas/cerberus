package permission_usecase

import (
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/pkg/order"
)

func getDefaultOrderBy() order.By {
	return order.NewBy("id", order.ASC)
}

func getOrderByFields() map[string]string {
	return map[string]string{
		"id":   permission.OrderByID,
		"name": permission.OrderByName,
	}
}
