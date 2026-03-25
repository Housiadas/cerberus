package user

import "github.com/Housiadas/cerberus/pkg/order"

// Set of fields that the results can be ordered by.
const (
	OrderByID      = "id"
	OrderByName    = "name"
	OrderByEmail   = "email"
	OrderByEnabled = "enabled"
)

func GetDefaultOrderBy() order.By {
	return order.NewBy(OrderByID, order.ASC)
}
