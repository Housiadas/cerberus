// Package role_permissions defines the write interface for the role_permissions join table.
package role_permissions

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist data.
type Storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (Storer, error)
	Add(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error
	Remove(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error
}
