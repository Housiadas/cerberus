package user_roles_permissions

import (
	"net/mail"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/google/uuid"
)

// Permission represents a permission with its ID and name.
type Permission struct {
	ID   uuid.UUID
	Name string
}

// UserRolesPermissions represents a single row from the user_roles_permissions view.
type UserRolesPermissions struct {
	userID         uuid.UUID
	userName       name.Name
	userEmail      mail.Address
	roleID         uuid.UUID
	roleName       name.Name
	permissionID   *uuid.UUID
	permissionName name.Null
}

// New constructs a UserRolesPermissions from its individual fields.
func New(
	userID uuid.UUID,
	userName name.Name,
	userEmail mail.Address,
	roleID uuid.UUID,
	roleName name.Name,
	permissionID *uuid.UUID,
	permissionName name.Null,
) UserRolesPermissions {
	return UserRolesPermissions{
		userID:         userID,
		userName:       userName,
		userEmail:      userEmail,
		roleID:         roleID,
		roleName:       roleName,
		permissionID:   permissionID,
		permissionName: permissionName,
	}
}

// UserID returns the user ID.
func (u UserRolesPermissions) UserID() uuid.UUID { return u.userID }

// UserName returns the user name.
func (u UserRolesPermissions) UserName() name.Name { return u.userName }

// UserEmail returns the user email.
func (u UserRolesPermissions) UserEmail() mail.Address { return u.userEmail }

// RoleID returns the role ID.
func (u UserRolesPermissions) RoleID() uuid.UUID { return u.roleID }

// RoleName returns the role name.
func (u UserRolesPermissions) RoleName() name.Name { return u.roleName }

// PermissionID returns the permission ID (nil if no permission).
func (u UserRolesPermissions) PermissionID() *uuid.UUID { return u.permissionID }

// PermissionName returns the permission name.
func (u UserRolesPermissions) PermissionName() name.Null { return u.permissionName }
