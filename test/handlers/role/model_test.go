package role_test

import (
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/internal/usecase/role_usecase"
	"github.com/Housiadas/cerberus/pkg/clock"
)

func toAppRole(r role.Role) role_usecase.Role {
	return role_usecase.Role{
		ID:        r.ID.String(),
		Name:      r.Name.String(),
		CreatedAt: clock.Format(&r.CreatedAt),
		UpdatedAt: clock.Format(&r.UpdatedAt),
	}
}

func toAppRoles(roles []role.Role) []role_usecase.Role {
	items := make([]role_usecase.Role, len(roles))
	for i, r := range roles {
		items[i] = toAppRole(r)
	}

	return items
}
