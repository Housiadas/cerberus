package tax_rate

import (
	"errors"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/country"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("tax rate not found")

type TaxRate struct {
	id          uuid.UUID
	name        string
	percentage  float64
	country     country.Country
	description string
	isActive    bool
	createdAt   time.Time
	updatedAt   time.Time
}

func New(
	id uuid.UUID,
	nm string,
	percentage float64,
	cntry country.Country,
	description string,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
) TaxRate {
	return TaxRate{
		id:          id,
		name:        nm,
		percentage:  percentage,
		country:     cntry,
		description: description,
		isActive:    isActive,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (t TaxRate) ID() uuid.UUID            { return t.id }
func (t TaxRate) Name() string             { return t.name }
func (t TaxRate) Percentage() float64      { return t.percentage }
func (t TaxRate) Country() country.Country { return t.country }
func (t TaxRate) Description() string      { return t.description }
func (t TaxRate) IsActive() bool           { return t.isActive }
func (t TaxRate) CreatedAt() time.Time     { return t.createdAt }
func (t TaxRate) UpdatedAt() time.Time     { return t.updatedAt }

func (t TaxRate) WithName(n string) TaxRate {
	t.name = n

	return t
}

func (t TaxRate) WithPercentage(p float64) TaxRate {
	t.percentage = p

	return t
}

func (t TaxRate) WithCountry(c country.Country) TaxRate {
	t.country = c

	return t
}

func (t TaxRate) WithDescription(d string) TaxRate {
	t.description = d

	return t
}

func (t TaxRate) WithIsActive(a bool) TaxRate {
	t.isActive = a

	return t
}

func (t TaxRate) WithUpdatedAt(tm time.Time) TaxRate {
	t.updatedAt = tm

	return t
}

type NewTaxRate struct {
	Name        string
	Percentage  float64
	Country     country.Country
	Description string
}

type UpdateTaxRate struct {
	Name        *string
	Percentage  *float64
	Country     *country.Country
	Description *string
	IsActive    *bool
}

type QueryFilter struct {
	ID       *uuid.UUID
	Name     *string
	Country  *country.Country
	IsActive *bool
}
