package audit_usecase

import (
	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/Housiadas/cerberus/pkg/clock"
)

// Audit represents information about an individual audit record.
type Audit struct {
	ID        string `json:"id"`
	ObjID     string `json:"objId"`
	ObjEntity string `json:"objEntity"`
	ObjName   string `json:"objName"`
	ActorID   string `json:"actorId"`
	Action    string `json:"action"`
	Data      string `json:"data"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type AuditPageResult struct {
	Data     []Audit       `json:"data"`
	Metadata page.Metadata `json:"metadata"`
}

func toAppAudit(aud audit.Audit) Audit {
	return Audit{
		ID:        aud.ID().String(),
		ObjID:     aud.ObjID().String(),
		ObjEntity: aud.ObjEntity().String(),
		ObjName:   aud.ObjName().String(),
		ActorID:   aud.ActorID().String(),
		Action:    aud.Action(),
		Data:      string(aud.Data()),
		Message:   aud.Message(),
		Timestamp: clock.Format(new(aud.Timestamp())),
	}
}

func toAppAudits(audits []audit.Audit) []Audit {
	app := make([]Audit, len(audits))
	for i, adt := range audits {
		app[i] = toAppAudit(adt)
	}

	return app
}
