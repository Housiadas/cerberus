package user_roles_permissions

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/google/uuid"
)

// Service manages the set of APIs for querying the user_roles_permissions view.
type Service struct {
	log    logger.Logger
	storer storer
}

// NewService constructs a business API for use.
func NewService(
	log logger.Logger,
	storer storer,
) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// QueryPermissionsByUserID returns all permissions (id and name) for the given user.
func (s *Service) QueryPermissionsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]Permission, error) {
	dbRows, err := s.storer.GetPermissionsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query permissions by user_id: %w", err)
	}

	return toDomainPermissions(dbRows), nil
}
