package role

import "github.com/Housiadas/cerberus/pkg/order"

// Set of fields that the results can be ordered by.
const (
	OrderByID   = "id"
	OrderByName = "name"
)

// GetDefaultOrderBy returns the default order for role queries.
func GetDefaultOrderBy() order.By {
	return order.NewBy(OrderByID, order.ASC)
}
