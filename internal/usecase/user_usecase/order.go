package user_usecase

import (
	"github.com/Housiadas/cerberus/internal/core/user"
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

func userFieldExtractor(orderBy order.By) func(User) any {
	return func(u User) any {
		switch orderBy.Field {
		case user.OrderByName:
			return u.Name
		case user.OrderByEmail:
			return u.Email
		case user.OrderByEnabled:
			return u.Enabled
		default:
			return u.ID
		}
	}
}
