package usage_record_repo

import (
	"github.com/Housiadas/cerberus/internal/core/domain/usage_record"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	usage_record.OrderByID:         "id",
	usage_record.OrderByRecordedAt: "recorded_at",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", ErrFieldNotExist
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
