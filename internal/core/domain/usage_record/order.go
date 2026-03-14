package usage_record

import "github.com/Housiadas/cerberus/pkg/order"

const (
	OrderByID         = "id"
	OrderByRecordedAt = "recorded_at"
)

func GetDefaultOrderBy() order.By {
	return order.NewBy(OrderByID, order.ASC)
}
