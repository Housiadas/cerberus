package audit_usecase

import (
	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/pkg/order"
)

var orderByFields = map[string]string{
	"obj_id":     audit.OrderByObjID,
	"obj_domain": audit.OrderByObjDomain,
	"obj_name":   audit.OrderByObjName,
	"actor_id":   audit.OrderByActorID,
	"action":     audit.OrderByAction,
}

func auditFieldExtractor(orderBy order.By) func(Audit) any {
	return func(a Audit) any {
		switch orderBy.Field {
		case audit.OrderByObjID:
			return a.ObjID
		case audit.OrderByObjDomain:
			return a.ObjEntity
		case audit.OrderByObjName:
			return a.ObjName
		case audit.OrderByActorID:
			return a.ActorID
		case audit.OrderByAction:
			return a.Action
		default:
			return a.ID
		}
	}
}
