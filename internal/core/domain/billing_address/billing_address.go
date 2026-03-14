package billing_address

import (
	"errors"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/country"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("billing address not found")

type BillingAddress struct {
	id         uuid.UUID
	accountID  uuid.UUID
	line1      string
	line2      string
	city       string
	state      string
	postalCode string
	country    country.Country
	createdAt  time.Time
	updatedAt  time.Time
}

func New(
	id uuid.UUID,
	accountID uuid.UUID,
	line1 string,
	line2 string,
	city string,
	state string,
	postalCode string,
	cntry country.Country,
	createdAt time.Time,
	updatedAt time.Time,
) BillingAddress {
	return BillingAddress{
		id: id, accountID: accountID, line1: line1, line2: line2,
		city: city, state: state, postalCode: postalCode, country: cntry,
		createdAt: createdAt, updatedAt: updatedAt,
	}
}

func (b BillingAddress) ID() uuid.UUID            { return b.id }
func (b BillingAddress) AccountID() uuid.UUID     { return b.accountID }
func (b BillingAddress) Line1() string            { return b.line1 }
func (b BillingAddress) Line2() string            { return b.line2 }
func (b BillingAddress) City() string             { return b.city }
func (b BillingAddress) State() string            { return b.state }
func (b BillingAddress) PostalCode() string       { return b.postalCode }
func (b BillingAddress) Country() country.Country { return b.country }
func (b BillingAddress) CreatedAt() time.Time     { return b.createdAt }
func (b BillingAddress) UpdatedAt() time.Time     { return b.updatedAt }

func (b BillingAddress) WithLine1(l string) BillingAddress {
	b.line1 = l

	return b
}

func (b BillingAddress) WithLine2(l string) BillingAddress {
	b.line2 = l

	return b
}

func (b BillingAddress) WithCity(c string) BillingAddress {
	b.city = c

	return b
}

func (b BillingAddress) WithState(s string) BillingAddress {
	b.state = s

	return b
}

func (b BillingAddress) WithPostalCode(p string) BillingAddress {
	b.postalCode = p

	return b
}

func (b BillingAddress) WithCountry(c country.Country) BillingAddress {
	b.country = c

	return b
}

func (b BillingAddress) WithUpdatedAt(t time.Time) BillingAddress {
	b.updatedAt = t

	return b
}

type NewBillingAddress struct {
	AccountID  uuid.UUID
	Line1      string
	Line2      string
	City       string
	State      string
	PostalCode string
	Country    country.Country
}

type UpdateBillingAddress struct {
	Line1      *string
	Line2      *string
	City       *string
	State      *string
	PostalCode *string
	Country    *country.Country
}
