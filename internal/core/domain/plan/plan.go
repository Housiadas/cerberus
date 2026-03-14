package plan

import (
	"errors"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("plan not found")

type Plan struct {
	id          uuid.UUID
	name        name.Name
	description string
	interval    Interval
	priceCents  int
	currency    currency.Currency
	isActive    bool
	createdAt   time.Time
	updatedAt   time.Time
}

func New(
	id uuid.UUID,
	nm name.Name,
	description string,
	interval Interval,
	priceCents int,
	cur currency.Currency,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
) Plan {
	return Plan{
		id:          id,
		name:        nm,
		description: description,
		interval:    interval,
		priceCents:  priceCents,
		currency:    cur,
		isActive:    isActive,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (p Plan) ID() uuid.UUID               { return p.id }
func (p Plan) Name() name.Name             { return p.name }
func (p Plan) Description() string         { return p.description }
func (p Plan) Interval() Interval          { return p.interval }
func (p Plan) PriceCents() int             { return p.priceCents }
func (p Plan) Currency() currency.Currency { return p.currency }
func (p Plan) IsActive() bool              { return p.isActive }
func (p Plan) CreatedAt() time.Time        { return p.createdAt }
func (p Plan) UpdatedAt() time.Time        { return p.updatedAt }

func (p Plan) WithName(n name.Name) Plan {
	p.name = n

	return p
}

func (p Plan) WithDescription(d string) Plan {
	p.description = d

	return p
}

func (p Plan) WithInterval(i Interval) Plan {
	p.interval = i

	return p
}

func (p Plan) WithPriceCents(c int) Plan {
	p.priceCents = c

	return p
}

func (p Plan) WithCurrency(c currency.Currency) Plan {
	p.currency = c

	return p
}

func (p Plan) WithIsActive(a bool) Plan {
	p.isActive = a

	return p
}

func (p Plan) WithUpdatedAt(t time.Time) Plan {
	p.updatedAt = t

	return p
}

type NewPlan struct {
	Name        name.Name
	Description string
	Interval    Interval
	PriceCents  int
	Currency    currency.Currency
}

type UpdatePlan struct {
	Name        *name.Name
	Description *string
	Interval    *Interval
	PriceCents  *int
	Currency    *currency.Currency
	IsActive    *bool
}

type QueryFilter struct {
	ID       *uuid.UUID
	Name     *name.Name
	IsActive *bool
}
