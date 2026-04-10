package role_permissions

import (
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toCreateRolePermissionParams(roleID, permissionID uuid.UUID) db.CreateRolePermissionParams {
	now := time.Now()

	return db.CreateRolePermissionParams{
		RoleID:       roleID,
		PermissionID: permissionID,
		CreatedAt:    pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:    pgtype.Timestamp{Time: now, Valid: true},
	}
}

func toDeleteRolePermissionParams(roleID, permissionID uuid.UUID) db.DeleteRolePermissionParams {
	return db.DeleteRolePermissionParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	}
}
