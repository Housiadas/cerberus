// Package coupon_service provides internal access to coupon core.
package coupon_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/coupon"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for coupon access.
type Service struct {
	log     logger.Logger
	storer  coupon.Storer
	uuidGen uuidgen.Generator
}

// New constructor.
func New(log logger.Logger, storer coupon.Storer, uuidGen uuidgen.Generator) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
	}
}

// NewWithTx constructs a new internal value that will use the
// specified transaction in any store-related calls.
func (c *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := c.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("coupon transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new coupon.Coupon to the system.
func (c *Service) Create(ctx context.Context, nc coupon.NewCoupon) (coupon.Coupon, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return coupon.Coupon{}, fmt.Errorf("coupon uuid generate: %w", err)
	}

	now := time.Now()
	cpn := coupon.New(id, nc.Code, nc.DiscountType, nc.DiscountValue, nc.Currency, nc.MaxRedemptions, 0, true, nc.ExpiresAt, now, now)

	err = c.storer.Create(ctx, cpn)
	if err != nil {
		return coupon.Coupon{}, fmt.Errorf("coupon create: %w", err)
	}

	return cpn, nil
}

// Update modifies information about a coupon.Coupon.
func (c *Service) Update(
	ctx context.Context,
	cpn coupon.Coupon,
	uc coupon.UpdateCoupon,
) (coupon.Coupon, error) {
	if uc.Code != nil {
		cpn = cpn.WithCode(*uc.Code)
	}
	if uc.DiscountType != nil {
		cpn = cpn.WithDiscountType(*uc.DiscountType)
	}
	if uc.DiscountValue != nil {
		cpn = cpn.WithDiscountValue(*uc.DiscountValue)
	}
	if uc.Currency != nil {
		cpn = cpn.WithCurrency(uc.Currency)
	}
	if uc.MaxRedemptions != nil {
		cpn = cpn.WithMaxRedemptions(uc.MaxRedemptions)
	}
	if uc.IsActive != nil {
		cpn = cpn.WithIsActive(*uc.IsActive)
	}
	if uc.ExpiresAt != nil {
		cpn = cpn.WithExpiresAt(uc.ExpiresAt)
	}

	cpn = cpn.WithUpdatedAt(time.Now())

	err := c.storer.Update(ctx, cpn)
	if err != nil {
		return coupon.Coupon{}, fmt.Errorf("coupon update: %w", err)
	}

	return cpn, nil
}

// Delete removes the specified coupon.Coupon.
func (c *Service) Delete(ctx context.Context, cpn coupon.Coupon) error {
	err := c.storer.Delete(ctx, cpn)
	if err != nil {
		return fmt.Errorf("coupon delete: %w", err)
	}

	return nil
}

// QueryByID finds the coupon by the specified ID.
func (c *Service) QueryByID(ctx context.Context, couponID uuid.UUID) (coupon.Coupon, error) {
	cpn, err := c.storer.QueryByID(ctx, couponID)
	if err != nil {
		return coupon.Coupon{}, fmt.Errorf("query: couponID[%s]: %w", couponID, err)
	}

	return cpn, nil
}

// QueryByCode finds the coupon by the specified code.
func (c *Service) QueryByCode(ctx context.Context, code string) (coupon.Coupon, error) {
	cpn, err := c.storer.QueryByCode(ctx, code)
	if err != nil {
		return coupon.Coupon{}, fmt.Errorf("query: code[%s]: %w", code, err)
	}

	return cpn, nil
}

// Query retrieves a list of existing coupons.
func (c *Service) Query(
	ctx context.Context,
	filter coupon.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]coupon.Coupon, error) {
	coupons, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("coupon query: %w", err)
	}

	return coupons, nil
}
