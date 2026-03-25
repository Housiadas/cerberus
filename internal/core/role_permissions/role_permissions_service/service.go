// Package role_permissions_service manages the set of APIs for role permission assignments.
package role_permissions_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/role_permissions"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Service manages role permission assignments.
type Service struct {
	log    logger.Logger
	storer role_permissions.Storer
}

// New constructs a business API for use.
func New(log logger.Logger, storer role_permissions.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new Service value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("role_permissions transaction issue: %w", err)
	}

	return &Service{
		log:    s.log,
		storer: storer,
	}, nil
}

// Add assigns a permission to a role.
func (s *Service) Add(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error {
	err := s.storer.Add(ctx, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("add role permission: %w", err)
	}

	return nil
}

// Remove removes a permission from a role.
func (s *Service) Remove(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error {
	err := s.storer.Remove(ctx, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("remove role permission: %w", err)
	}

	return nil
}
