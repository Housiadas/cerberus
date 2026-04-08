package user

import (
	"context"
	"net/mail"
	"time"

	"github.com/Housiadas/cerberus/internal/types/event"
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

// hasher defines the interface for password hashing operations.
type hasher interface {
	Hash(password string) ([]byte, error)
	Compare(hashedPassword []byte, password string) error
}

// dispatcher defines the interface for domain event dispatching.
type dispatcher interface {
	Dispatch(ctx context.Context, ev event.DomainEvent) error
}

// transactor defines the interface for transaction management.
type transactor interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	Create(ctx context.Context, usr User) error
	Update(ctx context.Context, usr User) error
	Delete(ctx context.Context, usr User) error
	Query(
		ctx context.Context,
		filter QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]User, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (User, error)
	QueryByEmail(ctx context.Context, email mail.Address) (User, error)
}
