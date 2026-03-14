package payment_method_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/payment_method"
	"github.com/google/uuid"
)

// TestNewPaymentMethods is a helper method for testing.
func TestNewPaymentMethods(n int) []payment_method.NewPaymentMethod {
	newPMs := make([]payment_method.NewPaymentMethod, n)

	for i := range n {
		npm := payment_method.NewPaymentMethod{
			AccountID: uuid.New(),
			Type:      payment_method.MustParse("card"),
			Details:   fmt.Sprintf("{\"last4\": \"%04d\"}", i),
			IsDefault: i == 0,
		}

		newPMs[i] = npm
	}

	return newPMs
}

// TestSeedPaymentMethods is a helper method for testing.
func TestSeedPaymentMethods(ctx context.Context, n int, service *Service) ([]payment_method.PaymentMethod, error) {
	newPMs := TestNewPaymentMethods(n)

	pms := make([]payment_method.PaymentMethod, len(newPMs))

	for i, npm := range newPMs {
		pm, err := service.Create(ctx, npm)
		if err != nil {
			return nil, fmt.Errorf("seeding payment method: idx: %d : %w", i, err)
		}

		pms[i] = pm
	}

	return pms, nil
}
