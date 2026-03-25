package account

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("account not found")
)

// Account represents a billing account (tenant).
type Account struct {
	id               uuid.UUID
	name             string
	stripeCustomerID sql.NullString
	createdAt        time.Time
	updatedAt        time.Time
	deletedAt        *time.Time
}

// New constructs an Account from its individual fields.
func New(
	id uuid.UUID,
	name string,
	stripeCustomerID sql.NullString,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) Account {
	return Account{
		id:               id,
		name:             name,
		stripeCustomerID: stripeCustomerID,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		deletedAt:        deletedAt,
	}
}

// ID returns the account ID.
func (a Account) ID() uuid.UUID { return a.id }

// Name returns the account name.
func (a Account) Name() string { return a.name }

// StripeCustomerID returns the Stripe customer ID.
func (a Account) StripeCustomerID() sql.NullString { return a.stripeCustomerID }

// CreatedAt returns the creation time.
func (a Account) CreatedAt() time.Time { return a.createdAt }

// UpdatedAt returns the last update time.
func (a Account) UpdatedAt() time.Time { return a.updatedAt }

// DeletedAt returns the deletion time (nil if not deleted).
func (a Account) DeletedAt() *time.Time { return a.deletedAt }

// WithName returns a new Account with the given name.
func (a Account) WithName(n string) Account {
	a.name = n

	return a
}

// WithStripeCustomerID returns a new Account with the given Stripe customer ID.
func (a Account) WithStripeCustomerID(id sql.NullString) Account {
	a.stripeCustomerID = id

	return a
}

// WithUpdatedAt returns a new Account with the given updated time.
func (a Account) WithUpdatedAt(t time.Time) Account {
	a.updatedAt = t

	return a
}

// MarshalJSON implements json.Marshaler.
func (a Account) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		ID               string `json:"id"`
		Name             string `json:"name"`
		StripeCustomerID string `json:"stripeCustomerId,omitempty"`
		CreatedAt        string `json:"createdAt"`
		UpdatedAt        string `json:"updatedAt"`
	}{
		ID:               a.id.String(),
		Name:             a.name,
		StripeCustomerID: a.stripeCustomerID.String,
		CreatedAt:        a.createdAt.UTC().Format(time.RFC3339),
		UpdatedAt:        a.updatedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal account: %w", err)
	}

	return data, nil
}

// NewAccount contains information needed to create a new account.
type NewAccount struct {
	Name string
}

// UpdateAccount contains information needed to update an account.
type UpdateAccount struct {
	Name             *string
	StripeCustomerID *sql.NullString
}

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	ID               *uuid.UUID
	Name             *string
	StripeCustomerID *string
}
