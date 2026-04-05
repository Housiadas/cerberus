// Package subscription is the service of the subscription domain.
package subscription

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/google/uuid"
)

// Service manages subscription operations.
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

// Create adds a new subscription to the system.
func (s *Service) Create(ctx context.Context, sub Subscription) error {
	err := s.storer.Create(ctx, sub)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

// Update modifies a subscription.
func (s *Service) Update(ctx context.Context, sub Subscription) error {
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
) ([]Subscription, error) {
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
) (Subscription, error) {
	sub, err := s.storer.QueryByStripeID(ctx, stripeSubID)
	if err != nil {
		return Subscription{}, fmt.Errorf("query by stripe id: %w", err)
	}

	return sub, nil
}
