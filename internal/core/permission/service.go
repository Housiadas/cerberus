// Package permission is the service of the permission domain
package permission

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Infoc(ctx context.Context, caller int, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Warnc(ctx context.Context, caller int, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Errorc(ctx context.Context, caller int, msg string, args ...any)
}

type generator interface {
	Generate() (uuid.UUID, error)
}

// dispatcher defines the interface for domain event dispatching.
type dispatcher interface {
	Dispatch(ctx context.Context, tran pgsql.CommitRollbacker, ev event.DomainEvent) error
}

// Service manages permission domain operations including persistence,
// transaction management, and event dispatching.
type Service struct {
	log        logger
	storer     Storer
	uuidGen    generator
	tx         pgsql.Beginner
	dispatcher dispatcher
}

// NewService constructor.
func NewService(
	log logger,
	storer Storer,
	uuidGen generator,
	tx pgsql.Beginner,
	dispatcher dispatcher,
) *Service {
	return &Service{
		log:        log,
		storer:     storer,
		uuidGen:    uuidGen,
		tx:         tx,
		dispatcher: dispatcher,
	}
}

// NewWithTx constructs a new Service that will use the
// specified transaction in any store-related calls.
func (s *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("permission transaction issue: %w", err)
	}

	svc := Service{
		log:        s.log,
		storer:     storer,
		uuidGen:    s.uuidGen,
		tx:         s.tx,
		dispatcher: s.dispatcher,
	}

	return &svc, nil
}

// Create adds a new Permission to the system within a transaction
// and dispatches a domain event.
func (s *Service) Create(
	ctx context.Context,
	np NewPermission,
) (Permission, error) {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return Permission{}, fmt.Errorf("permission uuid generate: %w", err)
	}

	now := time.Now()
	p := New(id, np.Name, now, now, nil)

	txErr := pgsql.RunInTx(ctx, s.log, s.tx, func(tran pgsql.CommitRollbacker) error {
		storerTx, err := s.storer.NewWithTx(tran)
		if err != nil {
			return fmt.Errorf("permission tx: %w", err)
		}

		if err = storerTx.Create(ctx, p); err != nil {
			return fmt.Errorf("create: %w", err)
		}

		return s.dispatcher.Dispatch(ctx, tran, newPermissionEvent(p, audit.ActionCreate))
	})
	if txErr != nil {
		return Permission{}, fmt.Errorf("create permission: %w", txErr)
	}

	return p, nil
}

// Update modifies information about a Permission within a transaction
// and dispatches a domain event.
func (s *Service) Update(
	ctx context.Context,
	p Permission,
	up UpdatePermission,
) (Permission, error) {
	if up.Name != nil {
		p = p.WithName(*up.Name)
	}

	p = p.WithUpdatedAt(time.Now())

	txErr := pgsql.RunInTx(ctx, s.log, s.tx, func(tran pgsql.CommitRollbacker) error {
		storerTx, err := s.storer.NewWithTx(tran)
		if err != nil {
			return fmt.Errorf("permission tx: %w", err)
		}

		if err = storerTx.Update(ctx, p); err != nil {
			return fmt.Errorf("update: %w", err)
		}

		return s.dispatcher.Dispatch(ctx, tran, newPermissionEvent(p, audit.ActionUpdate))
	})
	if txErr != nil {
		return Permission{}, fmt.Errorf("update permission: %w", txErr)
	}

	return p, nil
}

// Delete removes the specified Permission within a transaction
// and dispatches a domain event.
func (s *Service) Delete(ctx context.Context, p Permission) error {
	txErr := pgsql.RunInTx(ctx, s.log, s.tx, func(tran pgsql.CommitRollbacker) error {
		storerTx, err := s.storer.NewWithTx(tran)
		if err != nil {
			return fmt.Errorf("permission tx: %w", err)
		}

		if err = storerTx.Delete(ctx, p); err != nil {
			return fmt.Errorf("delete: %w", err)
		}

		return s.dispatcher.Dispatch(ctx, tran, newPermissionEvent(p, audit.ActionDelete))
	})
	if txErr != nil {
		return fmt.Errorf("delete permission: %w", txErr)
	}

	return nil
}

// QueryByID finds the permission by the specified ID.
func (s *Service) QueryByID(ctx context.Context, id uuid.UUID) (Permission, error) {
	p, err := s.storer.QueryByID(ctx, id)
	if err != nil {
		return Permission{}, fmt.Errorf("query: permissionID[%s]: %w", id, err)
	}

	return p, nil
}

// Query retrieves a list of existing permissions.
func (s *Service) Query(
	ctx context.Context,
	filter QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]Permission, error) {
	ps, err := s.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("permission query: %w", err)
	}

	return ps, nil
}

// newPermissionEvent creates a DomainEvent for permission operations.
func newPermissionEvent(p Permission, action string) event.DomainEvent {
	return event.DomainEvent{
		AggregateID: p.ID(),
		Payload:     p,
		ObjEntity:   entity.New(entity.PermissionEntity),
		ObjName:     p.Name(),
		Action:      action,
		Message:     "permission " + action,
	}
}
