package user_usecase

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"time"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/password"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

// =============================================================================

// AuthenticateUser defines the data needed to authenticate a user.
type AuthenticateUser struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Encode implements the encoder interface.
func (app *AuthenticateUser) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	if err != nil {
		return nil, web.ContentTypeJSON, fmt.Errorf("auth user encode error: %w", err)
	}

	return data, web.ContentTypeJSON, nil
}

// Validate checks the data in the model is considered clean.
func (app *AuthenticateUser) Validate() error {
	err := validation.Check(app)
	if err != nil {
		return fmt.Errorf("auth user validation error: %w", err)
	}

	return nil
}

// =============================================================================

// User represents information about an individual user.
type User struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"-"`
	Department   string `json:"department"`
	Enabled      bool   `json:"enabled"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// Encode implements the encoder interface.
func (app User) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	if err != nil {
		return nil, web.ContentTypeJSON, fmt.Errorf("user encode error: %w", err)
	}

	return data, web.ContentTypeJSON, nil
}

func toAppUser(bus user.User) User {
	return User{
		ID:           bus.ID.String(),
		Name:         bus.Name.String(),
		Email:        bus.Email.Address,
		PasswordHash: bus.PasswordHash,
		Department:   bus.Department.String(),
		Enabled:      bus.Enabled,
		CreatedAt:    bus.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    bus.UpdatedAt.Format(time.RFC3339),
	}
}

func toAppUsers(users []user.User) []User {
	app := make([]User, len(users))
	for i, usr := range users {
		app[i] = toAppUser(usr)
	}

	return app
}

// =============================================================================

type UserPageResult struct {
	Data     []User       `json:"data"`
	Metadata web.Metadata `json:"metadata"`
}

// =============================================================================

// NewUser defines the data needed to add a new user.
type NewUser struct {
	Name            string `json:"name"            validate:"required"`
	Email           string `json:"email"           validate:"required"`
	Department      string `json:"department"`
	Password        string `json:"password"        validate:"required"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required"`
}

// Decode implements the decoder interface.
func (app *NewUser) Decode(data []byte) error {
	err := json.Unmarshal(data, app)
	if err != nil {
		return fmt.Errorf("new user decode error: %w", err)
	}

	return nil
}

// Validate checks the data in the model is considered clean.
func (app *NewUser) Validate() error {
	err := validation.Check(app)
	if err != nil {
		return fmt.Errorf("new user validation error: %w", err)
	}

	return nil
}

func toBusNewUser(app NewUser) (user.NewUser, error) {
	var errors errs.FieldErrors

	addr, err := mail.ParseAddress(app.Email)
	if err != nil {
		errors.Add("email", err)
	}

	nme, err := name.Parse(app.Name)
	if err != nil {
		errors.Add("name", err)
	}

	department, err := name.ParseNull(app.Department)
	if err != nil {
		errors.Add("department", err)
	}

	pass, err := password.ParseConfirm(app.Password, app.PasswordConfirm)
	if err != nil {
		errors.Add("password", err)
	}

	if len(errors) > 0 {
		return user.NewUser{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	bus := user.NewUser{
		Name:       nme,
		Email:      *addr,
		Department: department,
		Password:   pass,
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
	err := json.Unmarshal(data, app)
	if err != nil {
		return fmt.Errorf("update user role decode error: %w", err)
	}

	return nil
}

// UpdateUser defines the data needed to update a user.
type UpdateUser struct {
	Name            *string `json:"name"`
	Email           *string `json:"email"`
	Department      *string `json:"department"`
	Password        *string `json:"password"`
	PasswordConfirm *string `json:"passwordConfirm"`
	Enabled         *bool   `json:"enabled"`
}

// Decode implements the decoder interface.
func (app *UpdateUser) Decode(data []byte) error {
	err := json.Unmarshal(data, app)
	if err != nil {
		return fmt.Errorf("update user decode error: %w", err)
	}

	return nil
}

func toBusUpdateUser(app UpdateUser) (user.UpdateUser, error) {
	var errors errs.FieldErrors

	var addr *mail.Address

	if app.Email != nil {
		var err error

		addr, err = mail.ParseAddress(*app.Email)
		if err != nil {
			errors.Add("email", err)
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

	var pass *password.Password

	p, err := password.ParseConfirmPointers(app.Password, app.PasswordConfirm)
	if err != nil {
		errors.Add("password", err)
	}

	pass = &p

	if len(errors) > 0 {
		return user.UpdateUser{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	bus := user.UpdateUser{
		Name:       nme,
		Email:      addr,
		Department: department,
		Password:   pass,
		Enabled:    app.Enabled,
	}

	return bus, nil
}
