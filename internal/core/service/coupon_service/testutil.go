package coupon_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/coupon"
)

// TestNewCoupons is a helper method for testing.
func TestNewCoupons(n int) []coupon.NewCoupon {
	newCoupons := make([]coupon.NewCoupon, n)

	for i := range n {
		nc := coupon.NewCoupon{
			Code:           fmt.Sprintf("COUPON%d", i),
			DiscountType:   coupon.MustParse("percent"),
			DiscountValue:  10,
			Currency:       nil,
			MaxRedemptions: nil,
			ExpiresAt:      nil,
		}

		newCoupons[i] = nc
	}

	return newCoupons
}

// TestSeedCoupons is a helper method for testing.
func TestSeedCoupons(ctx context.Context, n int, service *Service) ([]coupon.Coupon, error) {
	newCoupons := TestNewCoupons(n)

	coupons := make([]coupon.Coupon, len(newCoupons))

	for i, nc := range newCoupons {
		cpn, err := service.Create(ctx, nc)
		if err != nil {
			return nil, fmt.Errorf("seeding coupon: idx: %d : %w", i, err)
		}

		coupons[i] = cpn
	}

	return coupons, nil
}
