package permission_usecase

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

// Permission represents information about an individual permission.
type Permission struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"CreatedAt"`
	UpdatedAt string `json:"UpdatedAt"`
}

type PermissionPageResults struct {
	Data     []Permission `json:"data"`
	Metadata web.Metadata `json:"metadata"`
}

// Encode implements the encoder interface.
func (p Permission) Encode() ([]byte, string, error) {
	data, err := json.Marshal(p)
	return data, "application/json", err
}

func toAppPermission(p permission.Permission) Permission {
	return Permission{
		ID:        p.ID.String(),
		Name:      p.Name.String(),
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}

func toAppPermissions(perms []permission.Permission) []Permission {
	appPerms := make([]Permission, len(perms))
	for i, pr := range perms {
		appPerms[i] = toAppPermission(pr)
	}
	return appPerms
}

// NewPermission defines the data needed to add a new permission.
type NewPermission struct {
	Name string `json:"name" validate:"required"`
}

// Decode implements the decoder interface.
func (p *NewPermission) Decode(data []byte) error {
	return json.Unmarshal(data, p)
}

// Validate checks the data in the model is considered clean.
func (p *NewPermission) Validate() error {
	if err := validation.Check(p); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	return nil
}

func toBusNewPermission(app NewPermission) (permission.NewPermission, error) {
	nme, err := name.Parse(app.Name)
	if err != nil {
		return permission.NewPermission{}, fmt.Errorf("parse: %w", err)
	}
	return permission.NewPermission{
		Name: nme,
	}, nil
}

// UpdatePermission defines the data needed to update a permission.
type UpdatePermission struct {
	Name *string `json:"name"`
}

// Decode implements the decoder interface.
func (app *UpdatePermission) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app *UpdatePermission) Validate() error {
	if err := validation.Check(app); err != nil {
		return errs.Errorf(errs.InvalidArgument, "validation: %s", err)
	}
	return nil
}

func toBusUpdatePermission(app UpdatePermission) (permission.UpdatePermission, error) {
	var nme *name.Name
	if app.Name != nil {
		nm, err := name.Parse(*app.Name)
		if err != nil {
			return permission.UpdatePermission{}, fmt.Errorf("parse: %w", err)
		}
		nme = &nm
	}
	return permission.UpdatePermission{
		Name: nme,
	}, nil
}
