package user_roles_permissions_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/user_roles_permissions"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/uuid"
)

// Service manages the set of APIs for querying the user_roles_permissions view.
type Service struct {
	log    logger.Logger
	storer user_roles_permissions.Storer
}

// New constructs a business API for use.
func New(log logger.Logger, storer user_roles_permissions.Storer) *Service {
	return &Service{log: log, storer: storer}
}

// Query retrieves a list of user roles and permissions from the view.
func (s *Service) Query(
	ctx context.Context,
	filter user_roles_permissions.QueryFilter,
	orderBy order.By,
	p web.Page,
) ([]user_roles_permissions.UserRolesPermissions, error) {
	userRolesPerms, err := s.storer.Query(ctx, filter, orderBy, p)
	if err != nil {
		return nil, fmt.Errorf("user roles permissions query: %w", err)
	}

	return userRolesPerms, nil
}

// Count returns the total number of user roles and permissions that match the filter.
func (s *Service) Count(
	ctx context.Context,
	filter user_roles_permissions.QueryFilter,
) (int, error) {
	count, err := s.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("user roles permissions count: %w", err)
	}

	return count, nil
}

// HasPermission checks if the user has the specified permission.
func (s *Service) HasPermission(
	ctx context.Context,
	userID uuid.UUID,
	permissionName string,
) (bool, error) {
	hasPermissions, err := s.storer.HasPermission(ctx, userID, permissionName)
	if err != nil {
		return false, fmt.Errorf("user has permissions: %w", err)
	}

	return hasPermissions, nil
}
