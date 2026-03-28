// Package role is the service of the role domain
package role

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
	"github.com/jmoiron/sqlx"
)

type generator interface {
	Generate() (uuid.UUID, error)
}

type logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Infoc(ctx context.Context, caller int, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Warnc(ctx context.Context, caller int, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Errorc(ctx context.Context, caller int, msg string, args ...any)
}

// dispatcher defines the interface for domain event dispatching.
type dispatcher interface {
	Dispatch(ctx context.Context, ev event.DomainEvent) error
}

// Service manages role domain operations including persistence,
// transaction management, and event dispatching.
type Service struct {
	log        logger
	storer     Storer
	uuidGen    generator
	db         *sqlx.DB
	dispatcher dispatcher
}

// NewService constructor.
func NewService(
	log logger,
	storer Storer,
	uuidGen generator,
	db *sqlx.DB,
	dispatcher dispatcher,
) *Service {
	return &Service{
		log:        log,
		storer:     storer,
		uuidGen:    uuidGen,
		db:         db,
		dispatcher: dispatcher,
	}
}

// Create adds a new Role to the system within a transaction
// and dispatches a domain event.
func (c *Service) Create(ctx context.Context, nr NewRole) (Role, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return Role{}, fmt.Errorf("role uuid generate: %w", err)
	}

	now := time.Now()
	rol := New(id, nr.Name, now, now, nil)

	txErr := pgsql.RunInTx(ctx, c.log, c.db, func(txCtx context.Context) error {
		if err = c.storer.Create(txCtx, rol); err != nil {
			return fmt.Errorf("create: %w", err)
		}

		return c.dispatcher.Dispatch(txCtx, newRoleEvent(rol, audit.ActionCreate))
	})
	if txErr != nil {
		return Role{}, fmt.Errorf("create role: %w", txErr)
	}

	return rol, nil
}

// Update modifies information about a Role within a transaction
// and dispatches a domain event.
func (c *Service) Update(
	ctx context.Context,
	rl Role,
	uprole UpdateRole,
) (Role, error) {
	if uprole.Name != nil {
		rl = rl.WithName(*uprole.Name)
	}

	rl = rl.WithUpdatedAt(time.Now())

	txErr := pgsql.RunInTx(ctx, c.log, c.db, func(txCtx context.Context) error {
		if err := c.storer.Update(txCtx, rl); err != nil {
			return fmt.Errorf("update: %w", err)
		}

		return c.dispatcher.Dispatch(txCtx, newRoleEvent(rl, audit.ActionUpdate))
	})
	if txErr != nil {
		return Role{}, fmt.Errorf("update role: %w", txErr)
	}

	return rl, nil
}

// Delete removes the specified Role within a transaction
// and dispatches a domain event.
func (c *Service) Delete(ctx context.Context, rl Role) error {
	txErr := pgsql.RunInTx(ctx, c.log, c.db, func(txCtx context.Context) error {
		if err := c.storer.Delete(txCtx, rl); err != nil {
			return fmt.Errorf("delete: %w", err)
		}

		return c.dispatcher.Dispatch(txCtx, newRoleEvent(rl, audit.ActionDelete))
	})
	if txErr != nil {
		return fmt.Errorf("delete role: %w", txErr)
	}

	return nil
}

// QueryByID finds the role by the specified ID.
func (c *Service) QueryByID(ctx context.Context, roleID uuid.UUID) (Role, error) {
	rl, err := c.storer.QueryByID(ctx, roleID)
	if err != nil {
		return Role{}, fmt.Errorf("query: roleID[%s]: %w", roleID, err)
	}

	return rl, nil
}

// Query retrieves a list of existing roles.
func (c *Service) Query(
	ctx context.Context,
	filter QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]Role, error) {
	roles, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("role query: %w", err)
	}

	return roles, nil
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
