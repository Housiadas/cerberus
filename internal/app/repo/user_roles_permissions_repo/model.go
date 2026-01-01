package user_roles_permissions_repo

import (
	"database/sql"
	"fmt"
	"net/mail"

	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	urp "github.com/Housiadas/cerberus/internal/core/domain/user_roles_permissions"
)

type rowDB struct {
	UserID         uuid.UUID      `db:"user_id"`
	UserName       string         `db:"user_name"`
	UserEmail      string         `db:"user_email"`
	RoleID         uuid.UUID      `db:"role_id"`
	RoleName       string         `db:"role_name"`
	PermissionID   uuid.NullUUID  `db:"permission_id"`
	PermissionName sql.NullString `db:"permission_name"`
}

func toDomain(db rowDB) (urp.UserRolesPermissions, error) {
	userName, err := name.Parse(db.UserName)
	if err != nil {
		return urp.UserRolesPermissions{}, fmt.Errorf("parse user_name: %w", err)
	}
	roleName, err := name.Parse(db.RoleName)
	if err != nil {
		return urp.UserRolesPermissions{}, fmt.Errorf("parse role_name: %w", err)
	}
	permName, err := name.ParseNull(db.PermissionName.String)
	if err != nil {
		return urp.UserRolesPermissions{}, fmt.Errorf("parse permission_name: %w", err)
	}

	var permIDPtr *uuid.UUID
	if db.PermissionID.Valid {
		permID := db.PermissionID.UUID
		permIDPtr = &permID
	}

	return urp.UserRolesPermissions{
		UserID:         db.UserID,
		UserName:       userName,
		UserEmail:      mail.Address{Address: db.UserEmail},
		RoleID:         db.RoleID,
		RoleName:       roleName,
		PermissionID:   permIDPtr,
		PermissionName: permName,
	}, nil
}

func toDomains(dbs []rowDB) ([]urp.UserRolesPermissions, error) {
	res := make([]urp.UserRolesPermissions, len(dbs))
	for i, r := range dbs {
		dr, err := toDomain(r)
		if err != nil {
			return nil, err
		}
		res[i] = dr
	}
	return res, nil
}
