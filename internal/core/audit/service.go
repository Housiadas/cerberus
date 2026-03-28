package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	"github.com/google/uuid"
)

// Service manages the set of APIs for audit access.
type Service struct {
	log    logger.Logger
	storer Storer
}

// NewService constructs an audit business API for use.
func NewService(
	log logger.Logger,
	storer Storer,
) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// Create adds a new audit record to the system.
func (b *Service) Create(ctx context.Context, na NewAudit) (Audit, error) {
	ctx, span := telemetry.AddSpan(ctx, "business.auditbus.create")
	defer span.End()

	jsonData, err := json.Marshal(na.Data)
	if err != nil {
		return Audit{}, fmt.Errorf("marshal object: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Audit{}, fmt.Errorf("uuid: %w", err)
	}

	aud := New(
		id,
		na.ObjID,
		na.ObjEntity,
		na.ObjName,
		na.ActorID,
		na.Action,
		jsonData,
		na.Message,
		time.Now(),
	)

	err = b.storer.Create(ctx, aud)
	if err != nil {
		return Audit{}, fmt.Errorf("create audit: %w", err)
	}

	return aud, nil
}

// Query retrieves a list of existing audit records.
func (b *Service) Query(
	ctx context.Context,
	filter QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]Audit, error) {
	ctx, span := telemetry.AddSpan(ctx, "repo.audit.query")
	defer span.End()

	audits, err := b.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("query audits: %w", err)
	}

	return audits, nil
}
