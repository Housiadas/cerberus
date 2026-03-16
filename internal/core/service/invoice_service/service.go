// Package invoice_service is the service of the invoice domain
package invoice_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/invoice"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Service manages invoice operations.
type Service struct {
	log    logger.Logger
	storer invoice.Storer
}

// New constructs the service.
func New(log logger.Logger, storer invoice.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new Service that uses the specified transaction.
func (s *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("invoice transaction issue: %w", err)
	}

	return &Service{log: s.log, storer: storer}, nil
}

// Create adds a new invoice to the system.
func (s *Service) Create(ctx context.Context, inv invoice.Invoice) error {
	err := s.storer.Create(ctx, inv)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

// Update modifies an invoice.
func (s *Service) Update(ctx context.Context, inv invoice.Invoice) error {
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
) ([]invoice.Invoice, error) {
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
) (invoice.Invoice, error) {
	inv, err := s.storer.QueryByStripeID(ctx, stripeInvID)
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("query by stripe id: %w", err)
	}

	return inv, nil
}
