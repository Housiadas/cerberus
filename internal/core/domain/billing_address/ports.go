package billing_address

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, ba BillingAddress) error
	Update(ctx context.Context, ba BillingAddress) error
	QueryByID(ctx context.Context, billingAddressID uuid.UUID) (BillingAddress, error)
	QueryByAccountID(ctx context.Context, accountID uuid.UUID) (BillingAddress, error)
}
