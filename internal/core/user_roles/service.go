// Package user_roles_service manages the set of APIs for user role assignments.
package user_roles

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

// Service manages user role assignments.
type Service struct {
	log    logger
	storer Storer
}

// NewService constructs a business API for use.
func NewService(log logger, storer Storer) *Service {
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
		return nil, fmt.Errorf("user_roles transaction issue: %w", err)
	}

	return &Service{
		log:    s.log,
		storer: storer,
	}, nil
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
