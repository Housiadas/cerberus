package billing_address

import "github.com/Housiadas/cerberus/pkg/order"

const (
	OrderByID = "id"
)

func GetDefaultOrderBy() order.By {
	return order.NewBy(OrderByID, order.ASC)
}
