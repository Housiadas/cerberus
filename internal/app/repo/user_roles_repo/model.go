package user_roles_repo

import (
	"time"

	"github.com/google/uuid"

	ur "github.com/Housiadas/cerberus/internal/core/domain/user_roles"
)

type userRolesDB struct {
	UserID      uuid.UUID `db:"user_id"`
	RoleID      uuid.UUID `db:"role_id"`
	DateCreated time.Time `db:"date_created"`
}

func toUserRoleDB(ur ur.UserRole) userRolesDB {
	return userRolesDB{
		UserID:      ur.UserID,
		RoleID:      ur.RoleID,
		DateCreated: ur.DateCreated.UTC(),
	}
}

func toDomain(db userRolesDB) ur.UserRole {
	return ur.UserRole{
		UserID:      db.UserID,
		RoleID:      db.RoleID,
		DateCreated: db.DateCreated.In(time.Local),
	}
}

func toDomains(dbs []userRolesDB) []ur.UserRole {
	out := make([]ur.UserRole, 0, len(dbs))
	for _, r := range dbs {
		out = append(out, toDomain(r))
	}
	return out
}
