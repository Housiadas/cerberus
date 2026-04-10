package account

import (
	"context"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
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

// storer interface declares the behavior this package needs to persist and retrieve data.
type storer interface {
	CreateAccount(ctx context.Context, arg db.CreateAccountParams) (db.Account, error)
	UpdateAccount(ctx context.Context, arg db.UpdateAccountParams) (db.Account, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	QueryAccounts(
		ctx context.Context,
		filter db.AccountQueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]db.Account, error)
	GetAccountByID(ctx context.Context, id uuid.UUID) (db.GetAccountByIDRow, error)
}
