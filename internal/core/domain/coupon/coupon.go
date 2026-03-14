package coupon

import (
	"errors"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("coupon not found")

type Coupon struct {
	id             uuid.UUID
	code           string
	discountType   DiscountType
	discountValue  int
	currency       *currency.Currency
	maxRedemptions *int
	timesRedeemed  int
	isActive       bool
	expiresAt      *time.Time
	createdAt      time.Time
	updatedAt      time.Time
}

func New(
	id uuid.UUID,
	code string,
	dt DiscountType,
	discountValue int,
	cur *currency.Currency,
	maxRedemptions *int,
	timesRedeemed int,
	isActive bool,
	expiresAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) Coupon {
	return Coupon{
		id: id, code: code, discountType: dt, discountValue: discountValue,
		currency: cur, maxRedemptions: maxRedemptions, timesRedeemed: timesRedeemed,
		isActive: isActive, expiresAt: expiresAt, createdAt: createdAt, updatedAt: updatedAt,
	}
}

func (c Coupon) ID() uuid.UUID                { return c.id }
func (c Coupon) Code() string                 { return c.code }
func (c Coupon) DiscountType() DiscountType   { return c.discountType }
func (c Coupon) DiscountValue() int           { return c.discountValue }
func (c Coupon) Currency() *currency.Currency { return c.currency }
func (c Coupon) MaxRedemptions() *int         { return c.maxRedemptions }
func (c Coupon) TimesRedeemed() int           { return c.timesRedeemed }
func (c Coupon) IsActive() bool               { return c.isActive }
func (c Coupon) ExpiresAt() *time.Time        { return c.expiresAt }
func (c Coupon) CreatedAt() time.Time         { return c.createdAt }
func (c Coupon) UpdatedAt() time.Time         { return c.updatedAt }

func (c Coupon) WithCode(code string) Coupon {
	c.code = code

	return c
}

func (c Coupon) WithDiscountType(dt DiscountType) Coupon {
	c.discountType = dt

	return c
}

func (c Coupon) WithDiscountValue(v int) Coupon {
	c.discountValue = v

	return c
}

func (c Coupon) WithCurrency(cur *currency.Currency) Coupon {
	c.currency = cur

	return c
}

func (c Coupon) WithMaxRedemptions(m *int) Coupon {
	c.maxRedemptions = m

	return c
}

func (c Coupon) WithTimesRedeemed(t int) Coupon {
	c.timesRedeemed = t

	return c
}

func (c Coupon) WithIsActive(a bool) Coupon {
	c.isActive = a

	return c
}

func (c Coupon) WithExpiresAt(t *time.Time) Coupon {
	c.expiresAt = t

	return c
}

func (c Coupon) WithUpdatedAt(t time.Time) Coupon {
	c.updatedAt = t

	return c
}

type NewCoupon struct {
	Code           string
	DiscountType   DiscountType
	DiscountValue  int
	Currency       *currency.Currency
	MaxRedemptions *int
	ExpiresAt      *time.Time
}

type UpdateCoupon struct {
	Code           *string
	DiscountType   *DiscountType
	DiscountValue  *int
	Currency       *currency.Currency
	MaxRedemptions *int
	IsActive       *bool
	ExpiresAt      *time.Time
}

type QueryFilter struct {
	ID       *uuid.UUID
	Code     *string
	IsActive *bool
}
