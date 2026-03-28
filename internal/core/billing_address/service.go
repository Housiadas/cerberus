// Package billing_address_service is the service of the billing address domain
package billing_address

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

// Service manages billing address operations.
type Service struct {
	log    logger
	storer Storer
}

// NewService constructs the service.
func NewService(
	log logger,
	storer Storer,
) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// Create adds a new billing address to the system.
func (s *Service) Create(ctx context.Context, addr BillingAddress) error {
	err := s.storer.Create(ctx, addr)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

// Update modifies a billing address.
func (s *Service) Update(ctx context.Context, addr BillingAddress) error {
	err := s.storer.Update(ctx, addr)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

// Delete removes a billing address.
func (s *Service) Delete(ctx context.Context, addr BillingAddress) error {
	err := s.storer.Delete(ctx, addr)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// QueryByAccountID retrieves billing addresses for an account.
func (s *Service) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]BillingAddress, error) {
	addrs, err := s.storer.QueryByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("query by account id: %w", err)
	}

	return addrs, nil
}

// QueryByID retrieves a billing address by its ID.
func (s *Service) QueryByID(
	ctx context.Context,
	id uuid.UUID,
) (BillingAddress, error) {
	addr, err := s.storer.QueryByID(ctx, id)
	if err != nil {
		return BillingAddress{}, fmt.Errorf("query by id: %w", err)
	}

	return addr, nil
}
