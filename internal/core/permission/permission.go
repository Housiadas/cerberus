// Package permission represents the permission type in the system.
package permission

import (
	"errors"
	"time"

	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("permission not found")
	ErrRequiredName = errors.New("name is required")
)

// Permission represents information about an individual permission of our system.
type Permission struct {
	id        uuid.UUID
	name      name.Name
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

// New constructs a Permission from its individual fields.
func New(
	id uuid.UUID,
	nm name.Name,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) Permission {
	return Permission{
		id:        id,
		name:      nm,
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}
}

// ID returns the permission ID.
func (p Permission) ID() uuid.UUID { return p.id }

// Name returns the permission name.
func (p Permission) Name() name.Name { return p.name }

// CreatedAt returns the creation time.
func (p Permission) CreatedAt() time.Time { return p.createdAt }

// UpdatedAt returns the last update time.
func (p Permission) UpdatedAt() time.Time { return p.updatedAt }

// DeletedAt returns the deletion time (nil if not deleted).
func (p Permission) DeletedAt() *time.Time { return p.deletedAt }

// WithName returns a new Permission with the given name.
func (p Permission) WithName(n name.Name) Permission {
	p.name = n

	return p
}

// WithUpdatedAt returns a new Permission with the given updated time.
func (p Permission) WithUpdatedAt(t time.Time) Permission {
	p.updatedAt = t

	return p
}

// NewPermission contains information needed to create a new permission.
type NewPermission struct {
	Name name.Name
}

// UpdatePermission contains information needed to update a permission.
type UpdatePermission struct {
	Name *name.Name
}

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID   *uuid.UUID
	Name *name.Name
}
