package user_usecase

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"time"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/Housiadas/cerberus/pkg/page"
)

// =============================================================================

// Role represents information about an individual user.
type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DateCreated string `json:"dateCreated"`
	DateUpdated string `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (r Role) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

func toAppRole(r role.Role) Role {
	return Role{
		ID:          r.ID.String(),
		Name:        r.Name.String(),
		DateCreated: r.DateCreated.Format(time.RFC3339),
		DateUpdated: r.DateUpdated.Format(time.RFC3339),
	}
}

func toAppUsers(roles []role.Role) []Role {
	appRoles := make([]Role, len(roles))
	for i, rl := range roles {
		appRoles[i] = toAppRole(rl)
	}

	return appRoles
}

// =============================================================================

type RolePageResults struct {
	Data     []Role        `json:"data"`
	Metadata page.Metadata `json:"metadata"`
}

// =============================================================================

// NewUser defines the data needed to add a new user.
type NewUser struct {
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	Roles           []string `json:"roles" validate:"required"`
	Department      string   `json:"department"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"passwordConfirm" validate:"eqfield=Password"`
}

// Decode implements the decoder interface.
func (app *NewUser) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app *NewUser) Validate() error {
	if err := validation.Check(app); err != nil {
		return fmt.Errorf("validation: %w", err)

	}

	return nil
}

func toBusNewUser(app NewUser) (user.NewUser, error) {
	roles, err := role.ParseMany(app.Roles)
	if err != nil {
		return user.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	addr, err := mail.ParseAddress(app.Email)
	if err != nil {
		return user.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	nme, err := name.Parse(app.Name)
	if err != nil {
		return user.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	department, err := name.ParseNull(app.Department)
	if err != nil {
		return user.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	bus := user.NewUser{
		Name:       nme,
		Email:      *addr,
		RoleID:     roles,
		Department: department,
		Password:   app.Password,
	}

	return bus, nil
}

// =============================================================================

// UpdateUserRole defines the data needed to update a user role.
type UpdateUserRole struct {
	Roles []string `json:"roles" validate:"required"`
}

// Decode implements the decoder interface.
func (app *UpdateUserRole) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app *UpdateUserRole) Validate() error {
	if err := validation.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validation: %s", err)
	}

	return nil
}

func toBusUpdateUserRole(app UpdateUserRole) (user.UpdateUser, error) {
	var roles []role.Role
	if app.Roles != nil {
		roles = make([]role.Role, len(app.Roles))
		for i, roleStr := range app.Roles {
			r, err := role.Parse(roleStr)
			if err != nil {
				return user.UpdateUser{}, fmt.Errorf("parse: %w", err)
			}
			roles[i] = r
		}
	}

	bus := user.UpdateUser{
		RoleID: roles,
	}

	return bus, nil
}

// =============================================================================

// UpdateUser defines the data needed to update a user.
type UpdateUser struct {
	Name            *string `json:"name"`
	Email           *string `json:"email" validate:"omitempty,email"`
	Department      *string `json:"department"`
	Password        *string `json:"password"`
	PasswordConfirm *string `json:"passwordConfirm" validate:"omitempty,eqfield=Password"`
	Enabled         *bool   `json:"enabled"`
}

// Decode implements the decoder interface.
func (app *UpdateUser) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app *UpdateUser) Validate() error {
	if err := validation.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validation: %s", err)
	}

	return nil
}

func toBusUpdateUser(app UpdateUser) (user.UpdateUser, error) {
	var addr *mail.Address
	if app.Email != nil {
		var err error
		addr, err = mail.ParseAddress(*app.Email)
		if err != nil {
			return user.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
	}

	var nme *name.Name
	if app.Name != nil {
		nm, err := name.Parse(*app.Name)
		if err != nil {
			return user.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		nme = &nm
	}

	var department *name.Null
	if app.Department != nil {
		dep, err := name.ParseNull(*app.Department)
		if err != nil {
			return user.UpdateUser{}, fmt.Errorf("parse: %w", err)
		}
		department = &dep
	}

	bus := user.UpdateUser{
		Name:       nme,
		Email:      addr,
		Department: department,
		Password:   app.Password,
		Enabled:    app.Enabled,
	}

	return bus, nil
}
