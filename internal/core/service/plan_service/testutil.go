package plan_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/plan"
)

// TestNewPlans is a helper method for testing.
func TestNewPlans(n int) []plan.NewPlan {
	newPlans := make([]plan.NewPlan, n)

	for i := range n {
		np := plan.NewPlan{
			Name:        name.MustParse(fmt.Sprintf("Plan%d", i)),
			Description: fmt.Sprintf("Description%d", i),
			Interval:    plan.MustParse("monthly"),
			PriceCents:  1000,
			Currency:    currency.MustParse("USD"),
		}

		newPlans[i] = np
	}

	return newPlans
}

// TestSeedPlans is a helper method for testing.
func TestSeedPlans(ctx context.Context, n int, service *Service) ([]plan.Plan, error) {
	newPlans := TestNewPlans(n)

	plans := make([]plan.Plan, len(newPlans))

	for i, np := range newPlans {
		pl, err := service.Create(ctx, np)
		if err != nil {
			return nil, fmt.Errorf("seeding plan: idx: %d : %w", i, err)
		}

		plans[i] = pl
	}

	return plans, nil
}
