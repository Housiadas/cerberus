package user_roles_permissions

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

type logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Debugc(ctx context.Context, caller int, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Infoc(ctx context.Context, caller int, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Warnc(ctx context.Context, caller int, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Errorc(ctx context.Context, caller int, msg string, args ...any)
}

// Service manages the set of APIs for querying the user_roles_permissions view.
type Service struct {
	log    logger
	storer Storer
}

// NewService constructs a business API for use.
func NewService(
	log logger,
	storer Storer,
) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// Query retrieves a list of user roles and permissions from the view.
func (s *Service) Query(
	ctx context.Context,
	filter QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]UserRolesPermissions, error) {
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
) ([]Permission, error) {
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
