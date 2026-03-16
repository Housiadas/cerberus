// Package payment_service is the service of the payment domain
package payment_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/payment"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Service manages payment operations.
type Service struct {
	log    logger.Logger
	storer payment.Storer
}

// New constructs the service.
func New(log logger.Logger, storer payment.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new Service that uses the specified transaction.
func (s *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("payment transaction issue: %w", err)
	}

	return &Service{log: s.log, storer: storer}, nil
}

// Create adds a new payment to the system.
func (s *Service) Create(ctx context.Context, pmt payment.Payment) error {
	err := s.storer.Create(ctx, pmt)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

// Update modifies a payment.
func (s *Service) Update(ctx context.Context, pmt payment.Payment) error {
	err := s.storer.Update(ctx, pmt)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

// QueryByAccountID retrieves payments for an account.
func (s *Service) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]payment.Payment, error) {
	pmts, err := s.storer.QueryByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("query by account id: %w", err)
	}

	return pmts, nil
}

// QueryByStripeID retrieves a payment by its Stripe ID.
func (s *Service) QueryByStripeID(
	ctx context.Context,
	stripePaymentID string,
) (payment.Payment, error) {
	pmt, err := s.storer.QueryByStripeID(ctx, stripePaymentID)
	if err != nil {
		return payment.Payment{}, fmt.Errorf("query by stripe id: %w", err)
	}

	return pmt, nil
}
