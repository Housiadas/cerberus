package subscription_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/subscription"
	"github.com/google/uuid"
)

// TestNewSubscriptions is a helper method for testing.
func TestNewSubscriptions(n int) []subscription.NewSubscription {
	newSubscriptions := make([]subscription.NewSubscription, n)

	for i := range n {
		ns := subscription.NewSubscription{
			AccountID:          uuid.New(),
			PlanID:             uuid.New(),
			Status:             subscription.MustParse("active"),
			CurrentPeriodStart: time.Now(),
			CurrentPeriodEnd:   time.Now().Add(30 * 24 * time.Hour),
		}

		newSubscriptions[i] = ns
	}

	return newSubscriptions
}

// TestSeedSubscriptions is a helper method for testing.
func TestSeedSubscriptions(
	ctx context.Context,
	n int,
	service *Service,
) ([]subscription.Subscription, error) {
	newSubscriptions := TestNewSubscriptions(n)

	subscriptions := make([]subscription.Subscription, len(newSubscriptions))

	for i, ns := range newSubscriptions {
		sub, err := service.Create(ctx, ns)
		if err != nil {
			return nil, fmt.Errorf("seeding subscription: idx: %d : %w", i, err)
		}

		subscriptions[i] = sub
	}

	return subscriptions, nil
}
