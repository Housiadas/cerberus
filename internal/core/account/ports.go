package account

import (
	"context"
	"time"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}

// transactor defines the interface for transaction management.
type transactor interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	Create(ctx context.Context, acc Account) error
	Update(ctx context.Context, acc Account) error
	Delete(ctx context.Context, acc Account) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]Account, error)
	QueryByID(ctx context.Context, accountID uuid.UUID) (Account, error)
}
