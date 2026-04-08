package role_permissions

import (
	"context"

	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/google/uuid"
)

// dispatcher defines the interface for domain event dispatching.
type dispatcher interface {
	Dispatch(ctx context.Context, ev event.DomainEvent) error
}

// transactor defines the interface for transaction management.
type transactor interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// Storer interface declares the behavior this package needs to persist data.
type Storer interface {
	Add(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error
	Remove(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error
}
