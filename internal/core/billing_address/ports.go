package billing_address

import (
	"context"

	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	Create(ctx context.Context, addr BillingAddress) error
	Update(ctx context.Context, addr BillingAddress) error
	Delete(ctx context.Context, addr BillingAddress) error
	QueryByAccountID(ctx context.Context, accountID uuid.UUID) ([]BillingAddress, error)
	QueryByID(ctx context.Context, id uuid.UUID) (BillingAddress, error)
}
