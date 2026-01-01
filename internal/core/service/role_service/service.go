// Package role_service provides internal access to user core.
package role_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

// Service manages the set of APIs for user access.
type Service struct {
	log    *logger.Logger
	storer role.Storer
}

// New constructor
func New(log *logger.Logger, storer role.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new internal value that will use the
// specified transaction in any store-related calls.
func (c *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := c.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Service{
		log:    c.log,
		storer: storer,
	}

	return &bus, nil
}

// Create adds a new role.Role to the system.
func (c *Service) Create(ctx context.Context, nr role.NewRole) (role.Role, error) {
	now := time.Now()
	rol := role.Role{
		ID:        uuid.UUID{},
		Name:      nr.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := c.storer.Create(ctx, rol); err != nil {
		return role.Role{}, fmt.Errorf("role create: %w", err)
	}

	return rol, nil
}

// Update modifies information about a role.Role
func (c *Service) Update(ctx context.Context, rl role.Role, uprole role.UpdateRole) (role.Role, error) {
	if uprole.Name != nil {
		rl.Name = *uprole.Name
	}

	rl.UpdatedAt = time.Now()

	if err := c.storer.Update(ctx, rl); err != nil {
		return role.Role{}, fmt.Errorf("role update: %w", err)
	}

	return rl, nil
}

// Delete removes the specified role.Role
func (c *Service) Delete(ctx context.Context, rl role.Role) error {
	if err := c.storer.Delete(ctx, rl); err != nil {
		return fmt.Errorf("role delete: %w", err)
	}

	return nil
}

// Count returns the total number of users.
func (c *Service) Count(ctx context.Context, filter role.QueryFilter) (int, error) {
	return c.storer.Count(ctx, filter)
}

// QueryByID finds the user by the specified ID.
func (c *Service) QueryByID(ctx context.Context, roleID uuid.UUID) (role.Role, error) {
	rl, err := c.storer.QueryByID(ctx, roleID)
	if err != nil {
		return role.Role{}, fmt.Errorf("query: roleID[%s]: %w", roleID, err)
	}

	return rl, nil
}

// Query retrieves a list of existing users.
func (c *Service) Query(
	ctx context.Context,
	filter role.QueryFilter,
	orderBy order.By,
	page web.Page,
) ([]role.Role, error) {
	roles, err := c.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("role query: %w", err)
	}

	return roles, nil
}
