package role_permissions

import (
	"context"

	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist data.
type Storer interface {
	Add(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error
	Remove(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error
}
