package subscription_discount

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionDiscount struct {
	id             uuid.UUID
	subscriptionID uuid.UUID
	couponID       uuid.UUID
	appliedAt      time.Time
	createdAt      time.Time
}

func New(
	id uuid.UUID,
	subscriptionID uuid.UUID,
	couponID uuid.UUID,
	appliedAt time.Time,
	createdAt time.Time,
) SubscriptionDiscount {
	return SubscriptionDiscount{
		id: id, subscriptionID: subscriptionID, couponID: couponID,
		appliedAt: appliedAt, createdAt: createdAt,
	}
}

func (s SubscriptionDiscount) ID() uuid.UUID             { return s.id }
func (s SubscriptionDiscount) SubscriptionID() uuid.UUID { return s.subscriptionID }
func (s SubscriptionDiscount) CouponID() uuid.UUID       { return s.couponID }
func (s SubscriptionDiscount) AppliedAt() time.Time      { return s.appliedAt }
func (s SubscriptionDiscount) CreatedAt() time.Time      { return s.createdAt }

type NewSubscriptionDiscount struct {
	SubscriptionID uuid.UUID
	CouponID       uuid.UUID
}

type QueryFilter struct {
	ID             *uuid.UUID
	SubscriptionID *uuid.UUID
	CouponID       *uuid.UUID
}
