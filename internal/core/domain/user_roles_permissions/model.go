package user_roles_permissions

import (
	"net/mail"

	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
)

// UserRolesPermissions represents a single row from the user_roles_permissions view.
type UserRolesPermissions struct {
	UserID         uuid.UUID
	UserName       name.Name
	UserEmail      mail.Address
	RoleID         uuid.UUID
	RoleName       name.Name
	PermissionID   *uuid.UUID
	PermissionName name.Null
}
