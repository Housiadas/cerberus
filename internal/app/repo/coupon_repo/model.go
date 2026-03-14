package coupon_repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/coupon"
	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/google/uuid"
)

type couponDB struct {
	ID             uuid.UUID      `db:"id"`
	Code           string         `db:"code"`
	DiscountType   string         `db:"discount_type"`
	DiscountValue  int            `db:"discount_value"`
	Currency       sql.NullString `db:"currency"`
	MaxRedemptions sql.NullInt32  `db:"max_redemptions"`
	TimesRedeemed  int            `db:"times_redeemed"`
	IsActive       bool           `db:"is_active"`
	ExpiresAt      sql.NullTime   `db:"expires_at"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

func toCouponDB(cpn coupon.Coupon) couponDB {
	db := couponDB{
		ID:            cpn.ID(),
		Code:          cpn.Code(),
		DiscountType:  cpn.DiscountType().String(),
		DiscountValue: cpn.DiscountValue(),
		TimesRedeemed: cpn.TimesRedeemed(),
		IsActive:      cpn.IsActive(),
		ExpiresAt:     toNullTime(cpn.ExpiresAt()),
		CreatedAt:     cpn.CreatedAt().UTC(),
		UpdatedAt:     cpn.UpdatedAt().UTC(),
	}

	if cpn.Currency() != nil {
		db.Currency = sql.NullString{String: cpn.Currency().String(), Valid: true}
	}

	if cpn.MaxRedemptions() != nil {
		db.MaxRedemptions = sql.NullInt32{Int32: int32(*cpn.MaxRedemptions()), Valid: true}
	}

	return db
}

func toCouponDomain(db couponDB) (coupon.Coupon, error) {
	dt, err := coupon.Parse(db.DiscountType)
	if err != nil {
		return coupon.Coupon{}, fmt.Errorf("parse discount type: %w", err)
	}

	var cur *currency.Currency

	if db.Currency.Valid {
		c, err := currency.Parse(db.Currency.String)
		if err != nil {
			return coupon.Coupon{}, fmt.Errorf("parse currency: %w", err)
		}

		cur = &c
	}

	var maxRedemptions *int

	if db.MaxRedemptions.Valid {
		v := int(db.MaxRedemptions.Int32)
		maxRedemptions = &v
	}

	return coupon.New(
		db.ID,
		db.Code,
		dt,
		db.DiscountValue,
		cur,
		maxRedemptions,
		db.TimesRedeemed,
		db.IsActive,
		fromNullTime(db.ExpiresAt),
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	), nil
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: t.UTC(), Valid: true}
}

func fromNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}

	t := nt.Time.In(time.UTC)

	return &t
}

func toCouponsDomain(dbs []couponDB) ([]coupon.Coupon, error) {
	cpns := make([]coupon.Coupon, len(dbs))

	for i, db := range dbs {
		var err error

		cpns[i], err = toCouponDomain(db)
		if err != nil {
			return nil, fmt.Errorf("to coupons domain error: %w", err)
		}
	}

	return cpns, nil
}
