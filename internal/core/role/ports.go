package role

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

// dispatcher defines the interface for domain event dispatching.
type dispatcher interface {
	Dispatch(ctx context.Context, ev event.DomainEvent) error
}

// transactor defines the interface for transaction management.
type transactor interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type clock interface {
	Now() time.Time
}

// storer interface declares the behavior this package needs to persist and retrieve data.
type storer interface {
	CreateRole(ctx context.Context, arg db.CreateRoleParams) (db.Role, error)
	UpdateRole(ctx context.Context, arg db.UpdateRoleParams) (db.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	QueryRoles(
		ctx context.Context,
		filter db.RoleQueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]db.Role, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (db.GetRoleByIDRow, error)
}
