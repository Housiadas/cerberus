// Package billing_address_service is the service of the billing address domain
package billing_address_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/billing_address"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Service manages billing address operations.
type Service struct {
	log    logger.Logger
	storer billing_address.Storer
}

// New constructs the service.
func New(log logger.Logger, storer billing_address.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new Service that uses the specified transaction.
func (s *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("billing address transaction issue: %w", err)
	}

	return &Service{log: s.log, storer: storer}, nil
}

// Create adds a new billing address to the system.
func (s *Service) Create(ctx context.Context, addr billing_address.BillingAddress) error {
	err := s.storer.Create(ctx, addr)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

// Update modifies a billing address.
func (s *Service) Update(ctx context.Context, addr billing_address.BillingAddress) error {
	err := s.storer.Update(ctx, addr)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

// Delete removes a billing address.
func (s *Service) Delete(ctx context.Context, addr billing_address.BillingAddress) error {
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
) ([]billing_address.BillingAddress, error) {
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
) (billing_address.BillingAddress, error) {
	addr, err := s.storer.QueryByID(ctx, id)
	if err != nil {
		return billing_address.BillingAddress{}, fmt.Errorf("query by id: %w", err)
	}

	return addr, nil
}
