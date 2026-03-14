package refund_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/refund"
	"github.com/google/uuid"
)

// TestNewRefunds is a helper method for testing.
func TestNewRefunds(n int) []refund.NewRefund {
	newRefs := make([]refund.NewRefund, n)

	for i := range n {
		nref := refund.NewRefund{
			PaymentID:   uuid.New(),
			AmountCents: 1000 * (i + 1),
			Reason:      fmt.Sprintf("Reason%d", i),
		}

		newRefs[i] = nref
	}

	return newRefs
}

// TestSeedRefunds is a helper method for testing.
func TestSeedRefunds(ctx context.Context, n int, service *Service) ([]refund.Refund, error) {
	newRefs := TestNewRefunds(n)

	refs := make([]refund.Refund, len(newRefs))

	for i, nref := range newRefs {
		ref, err := service.Create(ctx, nref)
		if err != nil {
			return nil, fmt.Errorf("seeding refund: idx: %d : %w", i, err)
		}

		refs[i] = ref
	}

	return refs, nil
}
