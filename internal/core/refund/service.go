// Package refund_service is the service of the refund domain
package refund

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/google/uuid"
)

// Service manages refund operations.
type Service struct {
	log    logger.Logger
	storer Storer
}

// NewService constructs the service.
func NewService(
	log logger.Logger,
	storer Storer,
) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// Create adds a new refund to the system.
func (s *Service) Create(ctx context.Context, ref Refund) error {
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
) ([]Refund, error) {
	refs, err := s.storer.QueryByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("query by account id: %w", err)
	}

	return refs, nil
}
