package account

import (
	"errors"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("account not found")

type Account struct {
	id        uuid.UUID
	name      name.Name
	typ       Type
	enabled   bool
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

func New(
	id uuid.UUID,
	nm name.Name,
	typ Type,
	enabled bool,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) Account {
	return Account{
		id:        id,
		name:      nm,
		typ:       typ,
		enabled:   enabled,
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}
}

// ID returns the account ID.
func (a Account) ID() uuid.UUID { return a.id }

// Name returns the account name.
func (a Account) Name() name.Name { return a.name }

// Type returns the account type.
func (a Account) Type() Type { return a.typ }

// Enabled returns whether the account is enabled.
func (a Account) Enabled() bool { return a.enabled }

// CreatedAt returns the creation time.
func (a Account) CreatedAt() time.Time { return a.createdAt }

// UpdatedAt returns the last update time.
func (a Account) UpdatedAt() time.Time { return a.updatedAt }

// DeletedAt returns the deletion time.
func (a Account) DeletedAt() *time.Time { return a.deletedAt }

// WithName returns a new Account with the given name.
func (a Account) WithName(n name.Name) Account {
	a.name = n

	return a
}

// WithType returns a new Account with the given type.
func (a Account) WithType(t Type) Account {
	a.typ = t

	return a
}

// WithEnabled returns a new Account with the given enabled state.
func (a Account) WithEnabled(e bool) Account {
	a.enabled = e

	return a
}

// WithUpdatedAt returns a new Account with the given updated time.
func (a Account) WithUpdatedAt(t time.Time) Account {
	a.updatedAt = t

	return a
}

// NewAccount contains information needed to create a new account.
type NewAccount struct {
	Name name.Name
	Type Type
}

// UpdateAccount contains information needed to update an account.
type UpdateAccount struct {
	Name    *name.Name
	Type    *Type
	Enabled *bool
}

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID   *uuid.UUID
	Name *name.Name
	Type *Type
}
