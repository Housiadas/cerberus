package subscription

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, sub Subscription) error
	Update(ctx context.Context, sub Subscription) error
	QueryByAccountID(ctx context.Context, accountID uuid.UUID) ([]Subscription, error)
	QueryByStripeID(ctx context.Context, stripeSubID string) (Subscription, error)
}
