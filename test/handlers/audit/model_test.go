package audit_test

import (
	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/pkg/clock"
)

func toAppAudit(bus audit.Audit) audit_usecase.Audit {
	return audit_usecase.Audit{
		ID:        bus.ID.String(),
		ObjID:     bus.ObjID.String(),
		ObjEntity: bus.ObjEntity.String(),
		ObjName:   bus.ObjName.String(),
		ActorID:   bus.ActorID.String(),
		Action:    bus.Action,
		Data:      string(bus.Data),
		Message:   bus.Message,
		Timestamp: clock.Format(&bus.Timestamp),
	}
}

func toAppAudits(audits []audit.Audit) []audit_usecase.Audit {
	app := make([]audit_usecase.Audit, len(audits))
	for i, adt := range audits {
		app[i] = toAppAudit(adt)
	}

	return app
}
