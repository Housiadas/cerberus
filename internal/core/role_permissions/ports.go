package role_permissions

import (
	"context"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/types/event"
)

// dispatcher defines the interface for domain event dispatching.
type dispatcher interface {
	Dispatch(ctx context.Context, ev event.DomainEvent) error
}

// transactor defines the interface for transaction management.
type transactor interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// storer interface declares the behavior this package needs to persist data.
type storer interface {
	CreateRolePermission(
		ctx context.Context,
		arg db.CreateRolePermissionParams,
	) (db.RolePermission, error)
	DeleteRolePermission(ctx context.Context, arg db.DeleteRolePermissionParams) error
}
