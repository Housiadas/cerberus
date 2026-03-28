package payment

import (
	"context"

	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	Create(ctx context.Context, pmt Payment) error
	Update(ctx context.Context, pmt Payment) error
	QueryByAccountID(ctx context.Context, accountID uuid.UUID) ([]Payment, error)
	QueryByStripeID(ctx context.Context, stripePaymentID string) (Payment, error)
}
