package subscription_discount_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/subscription_discount"
	"github.com/google/uuid"
)

// TestNewSubscriptionDiscounts is a helper method for testing.
func TestNewSubscriptionDiscounts(n int) []subscription_discount.NewSubscriptionDiscount {
	newSDs := make([]subscription_discount.NewSubscriptionDiscount, n)

	for i := range n {
		nsd := subscription_discount.NewSubscriptionDiscount{
			SubscriptionID: uuid.New(),
			CouponID:       uuid.New(),
		}

		_ = i
		newSDs[i] = nsd
	}

	return newSDs
}

// TestSeedSubscriptionDiscounts is a helper method for testing.
func TestSeedSubscriptionDiscounts(ctx context.Context, n int, service *Service) ([]subscription_discount.SubscriptionDiscount, error) {
	newSDs := TestNewSubscriptionDiscounts(n)

	sds := make([]subscription_discount.SubscriptionDiscount, len(newSDs))

	for i, nsd := range newSDs {
		sd, err := service.Create(ctx, nsd)
		if err != nil {
			return nil, fmt.Errorf("seeding subscription discount: idx: %d : %w", i, err)
		}

		sds[i] = sd
	}

	return sds, nil
}
