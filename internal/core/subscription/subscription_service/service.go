// Package subscription_service is the service of the subscription domain
package subscription_service

import (
	"context"
	"fmt"

	subscription2 "github.com/Housiadas/cerberus/internal/core/subscription"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Service manages subscription operations.
type Service struct {
	log    logger.Logger
	storer subscription2.Storer
}

// New constructs the service.
func New(log logger.Logger, storer subscription2.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new Service that uses the specified transaction.
func (s *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("subscription transaction issue: %w", err)
	}

	return &Service{log: s.log, storer: storer}, nil
}

// Create adds a new subscription to the system.
func (s *Service) Create(ctx context.Context, sub subscription2.Subscription) error {
	err := s.storer.Create(ctx, sub)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

// Update modifies a subscription.
func (s *Service) Update(ctx context.Context, sub subscription2.Subscription) error {
	err := s.storer.Update(ctx, sub)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

// QueryByAccountID retrieves subscriptions for an account.
func (s *Service) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]subscription2.Subscription, error) {
	subs, err := s.storer.QueryByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("query by account id: %w", err)
	}

	return subs, nil
}

// QueryByStripeID retrieves a subscription by its Stripe ID.
func (s *Service) QueryByStripeID(
	ctx context.Context,
	stripeSubID string,
) (subscription2.Subscription, error) {
	sub, err := s.storer.QueryByStripeID(ctx, stripeSubID)
	if err != nil {
		return subscription2.Subscription{}, fmt.Errorf("query by stripe id: %w", err)
	}

	return sub, nil
}
