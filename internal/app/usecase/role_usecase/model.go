package role_usecase

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/pkg/web"
)

// =============================================================================

// Role represents information about an individual user.
type Role struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"CreatedAt"`
	UpdatedAt string `json:"UpdatedAt"`
}

type RolePageResult struct {
	Data     []Role       `json:"data"`
	Metadata web.Metadata `json:"metadata"`
}

// Encode implements the encoder interface.
func (r Role) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, "application/json", fmt.Errorf("role encode error: %w", err)
	}

	return data, "application/json", nil
}

func toAppRole(r role.Role) Role {
	return Role{
		ID:        r.ID.String(),
		Name:      r.Name.String(),
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
		UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
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

// Decode implements the decoder interface.
func (role *NewRole) Decode(data []byte) error {
	err := json.Unmarshal(data, role)
	if err != nil {
		return fmt.Errorf("new role decode error: %w", err)
	}

	return nil
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

// Decode implements the decoder interface.
func (app *UpdateRole) Decode(data []byte) error {
	err := json.Unmarshal(data, app)
	if err != nil {
		return fmt.Errorf("update role decode error: %w", err)
	}

	return nil
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
