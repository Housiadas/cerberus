package user_roles_permissions

import (
	"net/mail"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	UserID         *uuid.UUID
	UserName       *name.Name
	UserEmail      *mail.Address
	RoleID         *uuid.UUID
	RoleName       *name.Name
	PermissionID   *uuid.UUID
	PermissionName *name.Name
}
