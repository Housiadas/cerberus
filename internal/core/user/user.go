package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/internal/types/password"
	"github.com/google/uuid"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("user not found")
	ErrUniqueEmail           = errors.New("email is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrSearchUnavailable     = errors.New("search is not available")
)

// User represents information about an individual user.
type User struct {
	id           uuid.UUID
	name         name.Name
	email        mail.Address
	passwordHash []byte
	department   name.Null
	enabled      bool
	accountID    *uuid.UUID
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// New constructs a User from its individual fields.
func New(
	id uuid.UUID,
	nm name.Name,
	email mail.Address,
	passwordHash []byte,
	department name.Null,
	enabled bool,
	accountID *uuid.UUID,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) User {
	return User{
		id:           id,
		name:         nm,
		email:        email,
		passwordHash: passwordHash,
		department:   department,
		enabled:      enabled,
		accountID:    accountID,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}
}

// ID returns the user ID.
func (u User) ID() uuid.UUID { return u.id }

// Name returns the user name.
func (u User) Name() name.Name { return u.name }

// Email returns the user email.
func (u User) Email() mail.Address { return u.email }

// PasswordHash returns the password hash.
func (u User) PasswordHash() []byte { return u.passwordHash }

// Department returns the department.
func (u User) Department() name.Null { return u.department }

// Enabled returns whether the user is enabled.
func (u User) Enabled() bool { return u.enabled }

// AccountID returns the account ID.
func (u User) AccountID() *uuid.UUID { return u.accountID }

// CreatedAt returns the creation time.
func (u User) CreatedAt() time.Time { return u.createdAt }

// UpdatedAt returns the last update time.
func (u User) UpdatedAt() time.Time { return u.updatedAt }

// DeletedAt returns the deletion time (nil if not deleted).
func (u User) DeletedAt() *time.Time { return u.deletedAt }

// WithName returns a new User with the given name.
func (u User) WithName(n name.Name) User {
	u.name = n

	return u
}

// WithEmail returns a new User with the given email.
func (u User) WithEmail(e mail.Address) User {
	u.email = e

	return u
}

// WithPasswordHash returns a new User with the given password hash.
func (u User) WithPasswordHash(h []byte) User {
	u.passwordHash = h

	return u
}

// WithDepartment returns a new User with the given department.
func (u User) WithDepartment(d name.Null) User {
	u.department = d

	return u
}

// WithEnabled returns a new User with the given enabled state.
func (u User) WithEnabled(e bool) User {
	u.enabled = e

	return u
}

// WithAccountID returns a new User with the given account ID.
func (u User) WithAccountID(id *uuid.UUID) User {
	u.accountID = id

	return u
}

// WithUpdatedAt returns a new User with the given updated time.
func (u User) WithUpdatedAt(t time.Time) User {
	u.updatedAt = t

	return u
}

// MarshalJSON implements json.Marshaler.
func (u User) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Email      string `json:"email"`
		Department string `json:"department"`
		Enabled    bool   `json:"enabled"`
		CreatedAt  string `json:"createdAt"`
		UpdatedAt  string `json:"updatedAt"`
	}{
		ID:         u.id.String(),
		Name:       u.name.String(),
		Email:      u.email.Address,
		Department: u.department.String(),
		Enabled:    u.enabled,
		CreatedAt:  u.createdAt.UTC().Format(time.RFC3339),
		UpdatedAt:  u.updatedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal user: %w", err)
	}

	return data, nil
}

// NewUser contains information needed to create a new user.
type NewUser struct {
	Name       name.Name
	Email      mail.Address
	Department name.Null
	Password   password.Password
}

// UpdateUser contains information needed to update a user.
type UpdateUser struct {
	Name       *name.Name
	Email      *mail.Address
	Department *name.Null
	Password   *password.Password
	Enabled    *bool
}

// SearchResult holds the search results and pagination metadata.
type SearchResult struct {
	Users       []User
	TotalHits   int
	HasMore     bool
	SearchAfter string
}

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID               *uuid.UUID
	Name             *name.Name
	Email            *mail.Address
	StartCreatedDate *time.Time
	EndCreatedDate   *time.Time
}
