// Package user_roles_service manages the set of APIs for user role assignments.
package user_roles_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/user_roles"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/google/uuid"
)

// Service manages user role assignments.
type Service struct {
	log    logger.Logger
	storer user_roles.Storer
}

// New constructs a business API for use.
func New(log logger.Logger, storer user_roles.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// Add assigns a role to a user.
func (s *Service) Add(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	err := s.storer.Add(ctx, userID, roleID)
	if err != nil {
		return fmt.Errorf("add user role: %w", err)
	}

	return nil
}

// Remove removes a role from a user.
func (s *Service) Remove(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	err := s.storer.Remove(ctx, userID, roleID)
	if err != nil {
		return fmt.Errorf("remove user role: %w", err)
	}

	return nil
}
