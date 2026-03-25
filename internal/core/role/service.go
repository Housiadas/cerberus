// Package role_service provides internal access to user core.
package role

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type generator interface {
	Generate() (uuid.UUID, error)
}

type logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

// Service manages the set of APIs for user access.
type Service struct {
	log     logger
	storer  Storer
	uuidGen generator
}

// NewService constructor.
func NewService(
	log logger,
	storer Storer,
	uuidGen generator,
) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
	}
}

// NewWithTx constructs a new internal value that will use the
// specified transaction in any store-related calls.
func (c *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := c.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("role transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new role.Role to the system.
func (c *Service) Create(ctx context.Context, nr NewRole) (Role, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return Role{}, fmt.Errorf("role uuid generate: %w", err)
	}

	now := time.Now()
	rol := New(id, nr.Name, now, now, nil)

	err = c.storer.Create(ctx, rol)
	if err != nil {
		return Role{}, fmt.Errorf("role create: %w", err)
	}

	return rol, nil
}

// Update modifies information about a role.Role.
func (c *Service) Update(
	ctx context.Context,
	rl Role,
	uprole UpdateRole,
) (Role, error) {
	if uprole.Name != nil {
		rl = rl.WithName(*uprole.Name)
	}

	rl = rl.WithUpdatedAt(time.Now())

	err := c.storer.Update(ctx, rl)
	if err != nil {
		return Role{}, fmt.Errorf("role update: %w", err)
	}

	return rl, nil
}

// Delete removes the specified role.Role.
func (c *Service) Delete(ctx context.Context, rl Role) error {
	err := c.storer.Delete(ctx, rl)
	if err != nil {
		return fmt.Errorf("role delete: %w", err)
	}

	return nil
}

// QueryByID finds the user by the specified ID.
func (c *Service) QueryByID(ctx context.Context, roleID uuid.UUID) (Role, error) {
	rl, err := c.storer.QueryByID(ctx, roleID)
	if err != nil {
		return Role{}, fmt.Errorf("query: roleID[%s]: %w", roleID, err)
	}

	return rl, nil
}

// Query retrieves a list of existing users.
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
