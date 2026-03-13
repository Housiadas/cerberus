package role_usecase

import (
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/pkg/clock"
)

// =============================================================================

// Role represents information about an individual user.
type Role struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func toAppRole(r role.Role) Role {
	return Role{
		ID:        r.ID().String(),
		Name:      r.Name().String(),
		CreatedAt: clock.Format(new(r.CreatedAt())),
		UpdatedAt: clock.Format(new(r.UpdatedAt())),
	}
}

func toAppRoles(roles []role.Role) []Role {
	appRoles := make([]Role, len(roles))
	for i, rl := range roles {
		appRoles[i] = toAppRole(rl)
	}

	return appRoles
}

// =============================================================================

// NewRole defines the data needed to add a new user.
type NewRole struct {
	Name string `json:"name" validate:"required"`
}

func toBusNewRole(rl NewRole) (role.NewRole, error) {
	nme, err := name.Parse(rl.Name)
	if err != nil {
		return role.NewRole{}, fmt.Errorf("parse: %w", err)
	}

	return role.NewRole{
		Name: nme,
	}, nil
}

// =============================================================================

// UpdateRole defines the data needed to update a role.
type UpdateRole struct {
	Name *string `json:"name"`
}

func toBusUpdateUser(app UpdateRole) (role.UpdateRole, error) {
	var nme *name.Name

	if app.Name != nil {
		nm, err := name.Parse(*app.Name)
		if err != nil {
			return role.UpdateRole{}, fmt.Errorf("parse: %w", err)
		}

		nme = &nm
	}

	return role.UpdateRole{
		Name: nme,
	}, nil
}
