package plan_repo

import (
	"github.com/Housiadas/cerberus/internal/core/domain/plan"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	plan.OrderByID:   "id",
	plan.OrderByName: "name",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", ErrFieldNotExist
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
