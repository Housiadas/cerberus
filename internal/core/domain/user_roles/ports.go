// Package user_roles defines the write interface for the user_roles join table.
package user_roles

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist data.
type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Add(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
	Remove(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
}
