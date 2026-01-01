// Package role represents the role type in the system.
package role

import (
	"errors"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("role not found")
)

// Role represents information about an individual role of our system.
type Role struct {
	ID        uuid.UUID
	Name      name.Name
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NewRole struct {
	Name name.Name
}

type UpdateRole struct {
	Name *name.Name
}

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID   *uuid.UUID
	Name *name.Name
}
