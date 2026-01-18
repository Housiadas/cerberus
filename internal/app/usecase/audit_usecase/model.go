package audit_usecase

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/pkg/web"
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
	Data     []Audit      `json:"data"`
	Metadata web.Metadata `json:"metadata"`
}

// Encode implements the encoder interface.
func (app Audit) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	if err != nil {
		return nil, web.ContentTypeJSON, fmt.Errorf("audit encode error: %w", err)
	}

	return data, web.ContentTypeJSON, nil
}

func toAppAudit(aud audit.Audit) Audit {
	return Audit{
		ID:        aud.ID.String(),
		ObjID:     aud.ObjID.String(),
		ObjEntity: aud.ObjEntity.String(),
		ObjName:   aud.ObjName.String(),
		ActorID:   aud.ActorID.String(),
		Action:    aud.Action,
		Data:      string(aud.Data),
		Message:   aud.Message,
		Timestamp: aud.Timestamp.Format(time.RFC3339),
	}
}

func toAppAudits(audits []audit.Audit) []Audit {
	app := make([]Audit, len(audits))
	for i, adt := range audits {
		app[i] = toAppAudit(adt)
	}

	return app
}
