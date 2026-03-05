package permission_test

import (
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/internal/usecase/permission_usecase"
	"github.com/Housiadas/cerberus/pkg/clock"
)

func toAppPermission(p permission.Permission) permission_usecase.Permission {
	return permission_usecase.Permission{
		ID:        p.ID.String(),
		Name:      p.Name.String(),
		CreatedAt: clock.Format(&p.CreatedAt),
		UpdatedAt: clock.Format(&p.UpdatedAt),
	}
}

func toAppPermissions(perms []permission.Permission) []permission_usecase.Permission {
	items := make([]permission_usecase.Permission, len(perms))
	for i, p := range perms {
		items[i] = toAppPermission(p)
	}

	return items
}
