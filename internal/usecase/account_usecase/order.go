package account_usecase

import (
	"github.com/Housiadas/cerberus/internal/core/domain/account"
	"github.com/Housiadas/cerberus/pkg/order"
)

func getDefaultOrderBy() order.By {
	return order.NewBy("account_id", order.ASC)
}

func getOrderByFields() map[string]string {
	return map[string]string{
		"account_id": account.OrderByID,
		"name":       account.OrderByName,
	}
}

func accountFieldExtractor(orderBy order.By) func(Account) any {
	return func(a Account) any {
		switch orderBy.Field {
		case account.OrderByName:
			return a.Name
		default:
			return a.ID
		}
	}
}
