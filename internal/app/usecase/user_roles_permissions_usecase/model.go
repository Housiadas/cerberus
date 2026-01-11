package user_roles_permissions_usecase

import (
	urp "github.com/Housiadas/cerberus/internal/core/domain/user_roles_permissions"
)

// UserRolesPermissions represents a single row from the user_roles_permissions view for the app layer.
type UserRolesPermissions struct {
	UserID         string `json:"userId"`
	UserName       string `json:"userName"`
	UserEmail      string `json:"userEmail"`
	RoleID         string `json:"roleId"`
	RoleName       string `json:"roleName"`
	PermissionID   string `json:"permissionId,omitempty"`
	PermissionName string `json:"permissionName,omitempty"`
}

func toAppUserRolesPermissions(r urp.UserRolesPermissions) UserRolesPermissions {
	var permID string
	if r.PermissionID != nil {
		permID = r.PermissionID.String()
	}

	var permName string
	if r.PermissionName.Valid() {
		permName = r.PermissionName.String()
	}

	return UserRolesPermissions{
		UserID:         r.UserID.String(),
		UserName:       r.UserName.String(),
		UserEmail:      r.UserEmail.Address,
		RoleID:         r.RoleID.String(),
		RoleName:       r.RoleName.String(),
		PermissionID:   permID,
		PermissionName: permName,
	}
}

func toManyUserRolesPermissions(rows []urp.UserRolesPermissions) []UserRolesPermissions {
	res := make([]UserRolesPermissions, len(rows))
	for i, r := range rows {
		res[i] = toAppUserRolesPermissions(r)
	}

	return res
}
