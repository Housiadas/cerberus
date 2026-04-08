package permission

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

// Service manages permission domain operations including persistence,
// transaction management, and event dispatching.
type Service struct {
	log        logger.Logger
	storer     Storer
	uuidGen    generator
	tx         transactor
	dispatcher dispatcher
	clock      clock
}

// NewService constructor.
func NewService(
	log logger.Logger,
	storer Storer,
	uuidGen generator,
	tx transactor,
	dispatcher dispatcher,
	clock clock,
) *Service {
	return &Service{
		log:        log,
		storer:     storer,
		uuidGen:    uuidGen,
		tx:         tx,
		dispatcher: dispatcher,
		clock:      clock,
	}
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

	now := s.clock.Now()
	p := New(id, np.Name, now, now, nil)

	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		err = s.storer.Create(txCtx, p)
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}

		return s.dispatcher.Dispatch(txCtx, newPermissionEvent(p, audit.ActionCreate))
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

	p = p.WithUpdatedAt(s.clock.Now())

	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		err := s.storer.Update(txCtx, p)
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}

		return s.dispatcher.Dispatch(txCtx, newPermissionEvent(p, audit.ActionUpdate))
	})
	if txErr != nil {
		return Permission{}, fmt.Errorf("update permission: %w", txErr)
	}

	return p, nil
}

// Delete removes the specified Permission within a transaction
// and dispatches a domain event.
func (s *Service) Delete(ctx context.Context, p Permission) error {
	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		err := s.storer.Delete(txCtx, p)
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}

		return s.dispatcher.Dispatch(txCtx, newPermissionEvent(p, audit.ActionDelete))
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
