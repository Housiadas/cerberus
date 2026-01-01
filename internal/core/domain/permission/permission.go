// Package permission represents the permission type in the system.
package permission

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
)

var (
	ErrNotFound = errors.New("permission not found")
)

// Permission represents information about an individual permission of our system.
type Permission struct {
	ID        uuid.UUID
	Name      name.Name
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NewPermission struct {
	Name name.Name
}

type UpdatePermission struct {
	Name *name.Name
}

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID   *uuid.UUID
	Name *name.Name
}
