// Package refund_service is the service of the refund domain
package refund_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/refund"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Service manages refund operations.
type Service struct {
	log    logger.Logger
	storer refund.Storer
}

// New constructs the service.
func New(log logger.Logger, storer refund.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new Service that uses the specified transaction.
func (s *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("refund transaction issue: %w", err)
	}

	return &Service{log: s.log, storer: storer}, nil
}

// Create adds a new refund to the system.
func (s *Service) Create(ctx context.Context, ref refund.Refund) error {
	err := s.storer.Create(ctx, ref)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

// QueryByAccountID retrieves refunds for an account.
func (s *Service) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]refund.Refund, error) {
	refs, err := s.storer.QueryByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("query by account id: %w", err)
	}

	return refs, nil
}
