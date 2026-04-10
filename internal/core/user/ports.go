package user

import (
	"context"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
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

// storer interface declares the behavior this package needs to persist and retrieve data.
type storer interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	QueryUsers(
		ctx context.Context,
		filter db.UserQueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]db.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (db.GetUserByIDRow, error)
	GetUserByEmail(ctx context.Context, email string) (db.GetUserByEmailRow, error)
}
