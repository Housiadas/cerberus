package user_roles

import (
	"context"

	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist data.
type Storer interface {
	Add(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
	Remove(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
}
