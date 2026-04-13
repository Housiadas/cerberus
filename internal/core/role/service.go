package role

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

// Service manages role domain operations including persistence,
// transaction management, and event dispatching.
type Service struct {
	log        logger.Logger
	storer     storer
	uuidGen    generator
	tx         transactor
	dispatcher dispatcher
	clock      clock
}

// NewService constructor.
func NewService(
	log logger.Logger,
	storer storer,
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

// Create adds a new Role to the system within a transaction
// and dispatches a domain event.
func (s *Service) Create(ctx context.Context, nr NewRole) (Role, error) {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return Role{}, fmt.Errorf("role uuid generate: %w", err)
	}

	now := s.clock.Now()
	params := toCreateRoleParams(id, nr, now)

	var created Role

	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		dbRole, err := s.storer.CreateRole(txCtx, params)
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}

		created = toDomainRole(dbRole)

		return s.dispatcher.Dispatch(txCtx, newRoleEvent(created, audit.ActionCreate))
	})
	if txErr != nil {
		return Role{}, fmt.Errorf("create role: %w", txErr)
	}

	return created, nil
}

// Update modifies information about a Role within a transaction
// and dispatches a domain event.
func (s *Service) Update(
	ctx context.Context,
	rl Role,
	uprole UpdateRole,
) (Role, error) {
	if uprole.Name != nil {
		rl = rl.WithName(*uprole.Name)
	}

	rl = rl.WithUpdatedAt(s.clock.Now())

	params := toUpdateRoleParams(rl)

	var updated Role

	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		dbRole, err := s.storer.UpdateRole(txCtx, params)
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}

		updated = toDomainRole(dbRole)

		return s.dispatcher.Dispatch(txCtx, newRoleEvent(updated, audit.ActionUpdate))
	})
	if txErr != nil {
		return Role{}, fmt.Errorf("update role: %w", txErr)
	}

	return updated, nil
}

// Delete removes the specified Role within a transaction
// and dispatches a domain event.
func (s *Service) Delete(ctx context.Context, rl Role) error {
	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		err := s.storer.DeleteRole(txCtx, rl.ID())
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}

		return s.dispatcher.Dispatch(txCtx, newRoleEvent(rl, audit.ActionDelete))
	})
	if txErr != nil {
		return fmt.Errorf("delete role: %w", txErr)
	}

	return nil
}

// QueryByID finds the role by the specified ID.
func (s *Service) QueryByID(ctx context.Context, roleID uuid.UUID) (Role, error) {
	dbRole, err := s.storer.GetRoleByID(ctx, roleID)
	if err != nil {
		return Role{}, fmt.Errorf("query: roleID[%s]: %w", roleID, err)
	}

	return toDomainRole(dbRole), nil
}

// Query retrieves a list of existing roles.
func (s *Service) Query(
	ctx context.Context,
	filter QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]Role, error) {
	dbFilter := toDBQueryFilter(filter)

	dbRoles, err := s.storer.QueryRoles(ctx, dbFilter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("role query: %w", err)
	}

	return toDomainRoles(dbRoles), nil
}

// newRoleEvent creates a DomainEvent for role operations.
func newRoleEvent(rl Role, action string) event.DomainEvent {
	return event.DomainEvent{
		AggregateID: rl.ID(),
		Payload:     rl,
		ObjEntity:   entity.New(entity.RoleEntity),
		ObjName:     rl.Name(),
		Action:      action,
		Message:     "role " + action,
	}
}
