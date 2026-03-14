package coupon

import "github.com/Housiadas/cerberus/pkg/order"

const (
	OrderByID        = "id"
	OrderByCode      = "code"
	OrderByCreatedAt = "created_at"
)

func GetDefaultOrderBy() order.By {
	return order.NewBy(OrderByID, order.ASC)
}
