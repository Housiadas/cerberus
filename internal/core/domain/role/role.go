// Package role represents the role type in the system.
package role

import (
	"errors"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("role not found")

// Role represents information about an individual role of our system.
type Role struct {
	id        uuid.UUID
	name      name.Name
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

// New constructs a Role from its individual fields.
func New(
	id uuid.UUID,
	nm name.Name,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) Role {
	return Role{
		id:        id,
		name:      nm,
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}
}

// ID returns the role ID.
func (r Role) ID() uuid.UUID { return r.id }

// Name returns the role name.
func (r Role) Name() name.Name { return r.name }

// CreatedAt returns the creation time.
func (r Role) CreatedAt() time.Time { return r.createdAt }

// UpdatedAt returns the last update time.
func (r Role) UpdatedAt() time.Time { return r.updatedAt }

// DeletedAt returns the deletion time (nil if not deleted).
func (r Role) DeletedAt() *time.Time { return r.deletedAt }

// WithName returns a new Role with the given name.
func (r Role) WithName(n name.Name) Role {
	r.name = n

	return r
}

// WithUpdatedAt returns a new Role with the given updated time.
func (r Role) WithUpdatedAt(t time.Time) Role {
	r.updatedAt = t

	return r
}

// NewRole contains information needed to create a new role.
type NewRole struct {
	Name name.Name
}

// UpdateRole contains information needed to update a role.
type UpdateRole struct {
	Name *name.Name
}

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID   *uuid.UUID
	Name *name.Name
}
