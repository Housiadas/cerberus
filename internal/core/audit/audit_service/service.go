package audit_service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	audit2 "github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	"github.com/google/uuid"
)

// Service manages the set of APIs for audit access.
type Service struct {
	log    logger.Logger
	storer audit2.Storer
}

// New constructs an audit business API for use.
func New(log logger.Logger, storer audit2.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new Service value replacing the storer with a storer that is
// running within a transaction.
func (b *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("audit service tx: %w", err)
	}

	return &Service{
		log:    b.log,
		storer: storer,
	}, nil
}

// Create adds a new audit record to the system.
func (b *Service) Create(ctx context.Context, na audit2.NewAudit) (audit2.Audit, error) {
	ctx, span := telemetry.AddSpan(ctx, "business.auditbus.create")
	defer span.End()

	jsonData, err := json.Marshal(na.Data)
	if err != nil {
		return audit2.Audit{}, fmt.Errorf("marshal object: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return audit2.Audit{}, fmt.Errorf("uuid: %w", err)
	}

	aud := audit2.New(
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
		return audit2.Audit{}, fmt.Errorf("create audit: %w", err)
	}

	return aud, nil
}

// Query retrieves a list of existing audit records.
func (b *Service) Query(
	ctx context.Context,
	filter audit2.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]audit2.Audit, error) {
	ctx, span := telemetry.AddSpan(ctx, "repo.audit.query")
	defer span.End()

	audits, err := b.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("query audits: %w", err)
	}

	return audits, nil
}
