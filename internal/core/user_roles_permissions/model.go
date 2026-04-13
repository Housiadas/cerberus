package user_roles_permissions

import (
	"github.com/google/uuid"
)

// Permission represents a permission with its ID and name.
type Permission struct {
	ID   uuid.UUID
	Name string
}
