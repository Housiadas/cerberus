package user

import (
	"errors"
	"net/mail"
	"time"

	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("user not found")
	ErrUniqueEmail           = errors.New("email is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// User represents information about an individual user.
type User struct {
	ID           uuid.UUID
	RoleID       uuid.UUID
	Name         name.Name
	Email        mail.Address
	PasswordHash []byte
	Department   name.Null
	Enabled      bool
	DateCreated  time.Time
	DateUpdated  time.Time
}

// NewUser contains information needed to create a new user.
type NewUser struct {
	RoleID     uuid.UUID
	Name       name.Name
	Email      mail.Address
	Department name.Null
	Password   string
}

// UpdateUser contains information needed to update a user.
type UpdateUser struct {
	RoleID     uuid.UUID
	Name       *name.Name
	Email      *mail.Address
	Department *name.Null
	Password   *string
	Enabled    *bool
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
