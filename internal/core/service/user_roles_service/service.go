// Package user_roles_service provides internal access to user_roles core.
package user_roles_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/user_roles"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
	"github.com/google/uuid"
)

// Service manages the set of APIs for user access.
type Service struct {
	log    *logger.Logger
	storer user_roles.Storer
}

// New constructs a user.User internal API for use.
func New(log *logger.Logger, storer user_roles.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// Create adds a new role to the user.
func (c *Service) Create(ctx context.Context, nur user_roles.NewUserRole) (user_roles.UserRole, error) {
	now := time.Now()

	userRole := user_roles.UserRole{
		UserID:    nur.UserID,
		RoleID:    nur.RoleID,
		CreatedAt: now,
	}

	if err := c.storer.Create(ctx, userRole); err != nil {
		return user_roles.UserRole{}, fmt.Errorf("create: %w", err)
	}

	return userRole, nil
}

// Delete removes the specified user.
func (c *Service) Delete(ctx context.Context, usr user_roles.UserRole) error {
	if err := c.storer.Delete(ctx, usr); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users.
func (c *Service) Query(
	ctx context.Context,
	filter user_roles.QueryFilter,
	orderBy order.By,
	page page.Page,
) ([]user_roles.UserRole, error) {
	users, err := c.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return users, nil
}

// Count returns the total number of users.
func (c *Service) Count(ctx context.Context, filter user_roles.QueryFilter) (int, error) {
	return c.storer.Count(ctx, filter)
}

// GetUserRoleNames returns the roles of a user.
func (c *Service) GetUserRoleNames(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return c.storer.GetUserRoleNames(ctx, userID)
}
