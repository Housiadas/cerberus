// Package role_service provides internal access to user core.
package role_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Core manages the set of APIs for user access.
type Core struct {
	log    *logger.Logger
	storer role.Storer
}

// NewCore constructor
func NewCore(log *logger.Logger, storer role.Storer) *Core {
	return &Core{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new internal value that will use the
// specified transaction in any store-related calls.
func (c *Core) NewWithTx(tx pgsql.CommitRollbacker) (*Core, error) {
	storer, err := c.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Core{
		log:    c.log,
		storer: storer,
	}

	return &bus, nil
}

// Create adds a new role.Role to the system.
func (c *Core) Create(ctx context.Context, nr role.NewRole) (role.Role, error) {
	now := time.Now()
	rol := role.Role{
		ID:          uuid.UUID{},
		Name:        nr.Name,
		DateCreated: now,
		DateUpdated: now,
	}

	if err := c.storer.Create(ctx, rol); err != nil {
		return role.Role{}, fmt.Errorf("role create: %w", err)
	}

	return rol, nil
}

// Update modifies information about a role.Role
func (c *Core) Update(ctx context.Context, rl role.Role, uprole role.UpdateRole) (role.Role, error) {
	if uprole.Name != nil {
		rl.Name = *uprole.Name
	}

	rl.DateUpdated = time.Now()

	if err := c.storer.Update(ctx, rl); err != nil {
		return role.Role{}, fmt.Errorf("role update: %w", err)
	}

	return rl, nil
}

// Delete removes the specified role.Role
func (c *Core) Delete(ctx context.Context, rl role.Role) error {
	if err := c.storer.Delete(ctx, rl); err != nil {
		return fmt.Errorf("role delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users.
func (c *Core) Query(
	ctx context.Context,
	filter role.QueryFilter,
	orderBy order.By,
	page page.Page,
) ([]role.Role, error) {
	roles, err := c.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("role query: %w", err)
	}

	return roles, nil
}
