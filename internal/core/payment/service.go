// Package payment_service is the service of the payment domain
package payment

import (
	"context"
	"fmt"

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

// Service manages payment operations.
type Service struct {
	log    logger
	storer Storer
}

// NewService constructs the service.
func NewService(log logger, storer Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// Create adds a new payment to the system.
func (s *Service) Create(ctx context.Context, pmt Payment) error {
	err := s.storer.Create(ctx, pmt)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

// Update modifies a payment.
func (s *Service) Update(ctx context.Context, pmt Payment) error {
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
) ([]Payment, error) {
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
) (Payment, error) {
	pmt, err := s.storer.QueryByStripeID(ctx, stripePaymentID)
	if err != nil {
		return Payment{}, fmt.Errorf("query by stripe id: %w", err)
	}

	return pmt, nil
}
