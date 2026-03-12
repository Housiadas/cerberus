package user_usecase

import (
	"fmt"
	"net/mail"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/password"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/Housiadas/cerberus/internal/utils/validation"
	"github.com/Housiadas/cerberus/pkg/clock"
)

// =============================================================================

// AuthenticateUser defines the data needed to authenticate a user.
type AuthenticateUser struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required"`
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

func toAppUser(bus user.User) User {
	return User{
		ID:           bus.ID().String(),
		Name:         bus.Name().String(),
		Email:        bus.Email().Address,
		PasswordHash: bus.PasswordHash(),
		Department:   bus.Department().String(),
		Enabled:      bus.Enabled(),
		CreatedAt:    clock.Format(new(bus.CreatedAt())),
		UpdatedAt:    clock.Format(new(bus.UpdatedAt())),
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
	Data     []User        `json:"data"`
	Metadata page.Metadata `json:"metadata"`
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

// UpdateMe defines the data needed to update the authenticated user's own profile.
type UpdateMe struct {
	Name            *string `json:"name"`
	Email           *string `json:"email"`
	Department      *string `json:"department"`
	Password        *string `json:"password"`
	PasswordConfirm *string `json:"passwordConfirm"`
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

func toBusUpdateUser(app UpdateUser) (user.UpdateUser, error) {
	addr, pass, err := validateUpdateUserFields(app)
	if err != nil {
		return user.UpdateUser{}, err
	}

	var nme *name.Name

	if app.Name != nil {
		nm, nameErr := name.Parse(*app.Name)
		if nameErr != nil {
			return user.UpdateUser{}, fmt.Errorf("parse: %w", nameErr)
		}

		nme = &nm
	}

	var department *name.Null

	if app.Department != nil {
		dep, depErr := name.ParseNull(*app.Department)
		if depErr != nil {
			return user.UpdateUser{}, fmt.Errorf("parse: %w", depErr)
		}

		department = &dep
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

func validateUpdateUserFields(app UpdateUser) (*mail.Address, *password.Password, error) {
	var errors errs.FieldErrors

	var addr *mail.Address

	if app.Email != nil {
		var err error

		addr, err = mail.ParseAddress(*app.Email)
		if err != nil {
			errors.Add("email", err)
		}
	}

	var pass *password.Password

	if app.Password != nil || app.PasswordConfirm != nil {
		p, err := password.ParseConfirmPointers(app.Password, app.PasswordConfirm)
		if err != nil {
			errors.Add("password", err)
		}

		pass = &p
	}

	if len(errors) > 0 {
		return nil, nil, fmt.Errorf("validate: %w", errors.ToError())
	}

	return addr, pass, nil
}
