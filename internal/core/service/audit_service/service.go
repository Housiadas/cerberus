package audit_service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/otel"
)

// Service manages the set of APIs for audit access.
type Service struct {
	log    *logger.Logger
	storer audit.Storer
}

// New constructs an audit business API for use.
func New(log *logger.Logger, storer audit.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// Create adds a new audit record to the system.
func (b *Service) Create(ctx context.Context, na audit.NewAudit) (audit.Audit, error) {
	ctx, span := otel.AddSpan(ctx, "business.auditbus.create")
	defer span.End()

	jsonData, err := json.Marshal(na.Data)
	if err != nil {
		return audit.Audit{}, fmt.Errorf("marshal object: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return audit.Audit{}, fmt.Errorf("uuid: %w", err)
	}

	aud := audit.Audit{
		ID:        id,
		ObjID:     na.ObjID,
		ObjEntity: na.ObjEntity,
		ObjName:   na.ObjName,
		ActorID:   na.ActorID,
		Action:    na.Action,
		Data:      jsonData,
		Message:   na.Message,
		Timestamp: time.Now(),
	}

	if err := b.storer.Create(ctx, aud); err != nil {
		return audit.Audit{}, fmt.Errorf("create audit: %w", err)
	}

	return aud, nil
}

// Query retrieves a list of existing audit records.
func (b *Service) Query(ctx context.Context, filter audit.QueryFilter, orderBy order.By, page web.Page) ([]audit.Audit, error) {
	ctx, span := otel.AddSpan(ctx, "repo.audit.query")
	defer span.End()

	audits, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query audits: %w", err)
	}

	return audits, nil
}

// Count returns the total number of users.
func (b *Service) Count(ctx context.Context, filter audit.QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.auditbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}
