package user_usecase

import (
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/pkg/order"
)

func getDefaultOrderBy() order.By {
	return order.NewBy("user_id", order.ASC)
}

func getOrderByFields() map[string]string {
	return map[string]string{
		"user_id": user.OrderByID,
		"name":    user.OrderByName,
		"email":   user.OrderByEmail,
		"enabled": user.OrderByEnabled,
	}
}
