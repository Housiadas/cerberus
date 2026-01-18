package permission_repo

import (
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/pkg/order"
)

func orderByClause(orderBy order.By) (string, error) {
	by, exists := getOrderByFields()[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("%w: %s", errOrderFieldNotFound, orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}

func getOrderByFields() map[string]string {
	return map[string]string{
		permission.OrderByName: "name",
	}
}
