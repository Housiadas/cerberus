// Package invoice_service is the service of the invoice domain
package invoice

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/google/uuid"
)

// Service manages invoice operations.
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

// Create adds a new invoice to the system.
func (s *Service) Create(ctx context.Context, inv Invoice) error {
	err := s.storer.Create(ctx, inv)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

// Update modifies an invoice.
func (s *Service) Update(ctx context.Context, inv Invoice) error {
	err := s.storer.Update(ctx, inv)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

// QueryByAccountID retrieves invoices for an account.
func (s *Service) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]Invoice, error) {
	invs, err := s.storer.QueryByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("query by account id: %w", err)
	}

	return invs, nil
}

// QueryByStripeID retrieves an invoice by its Stripe ID.
func (s *Service) QueryByStripeID(
	ctx context.Context,
	stripeInvID string,
) (Invoice, error) {
	inv, err := s.storer.QueryByStripeID(ctx, stripeInvID)
	if err != nil {
		return Invoice{}, fmt.Errorf("query by stripe id: %w", err)
	}

	return inv, nil
}
