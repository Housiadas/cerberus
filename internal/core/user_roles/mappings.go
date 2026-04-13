package user_roles

import (
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toCreateUserRoleParams(userID, roleID uuid.UUID) db.CreateUserRoleParams {
	now := time.Now()

	return db.CreateUserRoleParams{
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: now, Valid: true},
	}
}

func toDeleteUserRoleParams(userID, roleID uuid.UUID) db.DeleteUserRoleParams {
	return db.DeleteUserRoleParams{
		UserID: userID,
		RoleID: roleID,
	}
}
