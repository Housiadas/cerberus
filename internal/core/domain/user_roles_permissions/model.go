package user_roles_permissions

import (
	"net/mail"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/google/uuid"
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
