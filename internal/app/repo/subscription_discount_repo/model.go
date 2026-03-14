package subscription_discount_repo

import (
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/subscription_discount"
	"github.com/google/uuid"
)

type subscriptionDiscountDB struct {
	ID             uuid.UUID `db:"id"`
	SubscriptionID uuid.UUID `db:"subscription_id"`
	CouponID       uuid.UUID `db:"coupon_id"`
	AppliedAt      time.Time `db:"applied_at"`
	CreatedAt      time.Time `db:"created_at"`
}

func toSubscriptionDiscountDB(
	sd subscription_discount.SubscriptionDiscount,
) subscriptionDiscountDB {
	return subscriptionDiscountDB{
		ID:             sd.ID(),
		SubscriptionID: sd.SubscriptionID(),
		CouponID:       sd.CouponID(),
		AppliedAt:      sd.AppliedAt().UTC(),
		CreatedAt:      sd.CreatedAt().UTC(),
	}
}

func toSubscriptionDiscountDomain(
	db subscriptionDiscountDB,
) subscription_discount.SubscriptionDiscount {
	return subscription_discount.New(
		db.ID,
		db.SubscriptionID,
		db.CouponID,
		db.AppliedAt.In(time.UTC),
		db.CreatedAt.In(time.UTC),
	)
}

func toSubscriptionDiscountsDomain(
	dbs []subscriptionDiscountDB,
) []subscription_discount.SubscriptionDiscount {
	sds := make([]subscription_discount.SubscriptionDiscount, len(dbs))

	for i, db := range dbs {
		sds[i] = toSubscriptionDiscountDomain(db)
	}

	return sds
}
