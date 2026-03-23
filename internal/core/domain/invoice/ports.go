package invoice

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, inv Invoice) error
	Update(ctx context.Context, inv Invoice) error
	QueryByAccountID(ctx context.Context, accountID uuid.UUID) ([]Invoice, error)
	QueryByStripeID(ctx context.Context, stripeInvID string) (Invoice, error)
}
