package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	"github.com/google/uuid"
)

type logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Debugc(ctx context.Context, caller int, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Infoc(ctx context.Context, caller int, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Warnc(ctx context.Context, caller int, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Errorc(ctx context.Context, caller int, msg string, args ...any)
}

// Service manages the set of APIs for audit access.
type Service struct {
	log    logger
	storer Storer
}

// NewService constructs an audit business API for use.
func NewService(
	log logger,
	storer Storer,
) *Service {
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
