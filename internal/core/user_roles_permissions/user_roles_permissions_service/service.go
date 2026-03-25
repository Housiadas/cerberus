package user_roles_permissions_service

import (
	"context"
	"fmt"

	user_roles_permissions2 "github.com/Housiadas/cerberus/internal/core/user_roles_permissions"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

// Service manages the set of APIs for querying the user_roles_permissions view.
type Service struct {
	log    logger.Logger
	storer user_roles_permissions2.Storer
}

// New constructs a business API for use.
func New(log logger.Logger, storer user_roles_permissions2.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// Query retrieves a list of user roles and permissions from the view.
func (s *Service) Query(
	ctx context.Context,
	filter user_roles_permissions2.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]user_roles_permissions2.UserRolesPermissions, error) {
	userRolesPerms, err := s.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("user roles permissions query: %w", err)
	}

	return userRolesPerms, nil
}

// QueryPermissionsByUserID returns all permissions (id and name) for the given user.
func (s *Service) QueryPermissionsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]user_roles_permissions2.Permission, error) {
	permissions, err := s.storer.QueryPermissionsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query permissions by user_id: %w", err)
	}

	return permissions, nil
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
