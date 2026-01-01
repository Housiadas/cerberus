// Package user_roles represents the relation between users and roles.
package user_roles

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("user roles not found")
)

// UserRole represents a single assignment of a role to a user.
type UserRole struct {
	UserID    uuid.UUID
	RoleID    uuid.UUID
	CreatedAt time.Time
}

type NewUserRole struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}

// QueryFilter holds the available fields a query can be filtered on.
// We use pointer semantics to allow distinguishing zero values from non-provided values.
type QueryFilter struct {
	UserID *uuid.UUID
	RoleID *uuid.UUID
}
