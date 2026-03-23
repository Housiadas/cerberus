package billing_address

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("billing address not found")

// BillingAddress represents a billing address for an account.
type BillingAddress struct {
	id         uuid.UUID
	accountID  uuid.UUID
	line1      string
	line2      string
	city       string
	state      string
	postalCode string
	country    string
	createdAt  time.Time
	updatedAt  time.Time
}

// New constructs a BillingAddress from its individual fields.
func New(
	id uuid.UUID,
	accountID uuid.UUID,
	line1 string,
	line2 string,
	city string,
	state string,
	postalCode string,
	country string,
	createdAt time.Time,
	updatedAt time.Time,
) BillingAddress {
	return BillingAddress{
		id:         id,
		accountID:  accountID,
		line1:      line1,
		line2:      line2,
		city:       city,
		state:      state,
		postalCode: postalCode,
		country:    country,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

// ID returns the billing address ID.
func (b BillingAddress) ID() uuid.UUID { return b.id }

// AccountID returns the account ID.
func (b BillingAddress) AccountID() uuid.UUID { return b.accountID }

// Line1 returns address line 1.
func (b BillingAddress) Line1() string { return b.line1 }

// Line2 returns address line 2.
func (b BillingAddress) Line2() string { return b.line2 }

// City returns the city.
func (b BillingAddress) City() string { return b.city }

// State returns the state.
func (b BillingAddress) State() string { return b.state }

// PostalCode returns the postal code.
func (b BillingAddress) PostalCode() string { return b.postalCode }

// Country returns the country code.
func (b BillingAddress) Country() string { return b.country }

// CreatedAt returns the creation time.
func (b BillingAddress) CreatedAt() time.Time { return b.createdAt }

// UpdatedAt returns the last update time.
func (b BillingAddress) UpdatedAt() time.Time { return b.updatedAt }

// WithLine1 returns a new BillingAddress with the given line1.
func (b BillingAddress) WithLine1(l string) BillingAddress {
	b.line1 = l

	return b
}

// WithLine2 returns a new BillingAddress with the given line2.
func (b BillingAddress) WithLine2(l string) BillingAddress {
	b.line2 = l

	return b
}

// WithCity returns a new BillingAddress with the given city.
func (b BillingAddress) WithCity(c string) BillingAddress {
	b.city = c

	return b
}

// WithState returns a new BillingAddress with the given state.
func (b BillingAddress) WithState(s string) BillingAddress {
	b.state = s

	return b
}

// WithPostalCode returns a new BillingAddress with the given postal code.
func (b BillingAddress) WithPostalCode(p string) BillingAddress {
	b.postalCode = p

	return b
}

// WithCountry returns a new BillingAddress with the given country.
func (b BillingAddress) WithCountry(c string) BillingAddress {
	b.country = c

	return b
}

// WithUpdatedAt returns a new BillingAddress with the given updated time.
func (b BillingAddress) WithUpdatedAt(t time.Time) BillingAddress {
	b.updatedAt = t

	return b
}
