package user_roles_usecase

import (
	"time"

	ur "github.com/Housiadas/cerberus/internal/core/domain/user_roles"
)

// UserRole represents a single row from the user_roles view for the app layer.
type UserRole struct {
	UserID    string    `json:"user_id"`
	RoleID    string    `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

func toAppUserRolesPermissions(usrRole ur.UserRole) UserRole {
	return UserRole{
		UserID:    usrRole.UserID.String(),
		RoleID:    usrRole.RoleID.String(),
		CreatedAt: usrRole.CreatedAt.UTC(),
	}
}

func toManyUserRolesPermissions(rows []ur.UserRole) []UserRole {
	res := make([]UserRole, len(rows))
	for i, r := range rows {
		res[i] = toAppUserRolesPermissions(r)
	}
	return res
}
