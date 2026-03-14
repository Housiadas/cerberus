package tax_rate

import "github.com/Housiadas/cerberus/pkg/order"

const (
	OrderByID   = "id"
	OrderByName = "name"
)

func GetDefaultOrderBy() order.By {
	return order.NewBy(OrderByID, order.ASC)
}
