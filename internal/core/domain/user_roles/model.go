// Package user_roles represents the relation between users and roles.
package user_roles

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents a single assignment of a role to a user.
type UserRole struct {
	UserID      uuid.UUID
	RoleID      uuid.UUID
	DateCreated time.Time
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
