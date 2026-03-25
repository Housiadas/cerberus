package audit_repo

import (
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
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
		audit.OrderByObjID:     "obj_id",
		audit.OrderByObjDomain: "obj_domain",
		audit.OrderByObjName:   "obj_name",
		audit.OrderByActorID:   "actor_id",
		audit.OrderByAction:    "action",
	}
}
