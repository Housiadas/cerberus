package user_roles_permissions

import (
	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/google/uuid"
)

func toDomainPermissions(rows []db.GetPermissionsByUserIDRow) []Permission {
	perms := make([]Permission, len(rows))
	for i, r := range rows {
		var id uuid.UUID
		if r.PermissionID.Valid {
			id = r.PermissionID.Bytes
		}

		perms[i] = Permission{
			ID:   id,
			Name: r.PermissionName.String,
		}
	}

	return perms
}
