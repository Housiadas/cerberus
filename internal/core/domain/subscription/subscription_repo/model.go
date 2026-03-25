package subscription_repo

import (
	"database/sql"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/subscription"
	"github.com/google/uuid"
)

type subscriptionDB struct {
	ID                   uuid.UUID    `db:"id"`
	AccountID            uuid.UUID    `db:"account_id"`
	StripeSubscriptionID string       `db:"stripe_subscription_id"`
	StripeCustomerID     string       `db:"stripe_customer_id"`
	StripePriceID        string       `db:"stripe_price_id"`
	Status               string       `db:"status"`
	CurrentPeriodStart   sql.NullTime `db:"current_period_start"`
	CurrentPeriodEnd     sql.NullTime `db:"current_period_end"`
	CancelAtPeriodEnd    bool         `db:"cancel_at_period_end"`
	CanceledAt           sql.NullTime `db:"canceled_at"`
	CreatedAt            time.Time    `db:"created_at"`
	UpdatedAt            time.Time    `db:"updated_at"`
}

func toSubscriptionDB(sub subscription.Subscription) subscriptionDB {
	return subscriptionDB{
		ID:                   sub.ID(),
		AccountID:            sub.AccountID(),
		StripeSubscriptionID: sub.StripeSubscriptionID(),
		StripeCustomerID:     sub.StripeCustomerID(),
		StripePriceID:        sub.StripePriceID(),
		Status:               sub.Status(),
		CurrentPeriodStart:   toNullTime(sub.CurrentPeriodStart()),
		CurrentPeriodEnd:     toNullTime(sub.CurrentPeriodEnd()),
		CancelAtPeriodEnd:    sub.CancelAtPeriodEnd(),
		CanceledAt:           toNullTime(sub.CanceledAt()),
		CreatedAt:            sub.CreatedAt().UTC(),
		UpdatedAt:            sub.UpdatedAt().UTC(),
	}
}

func toSubscriptionDomain(db subscriptionDB) subscription.Subscription {
	return subscription.New(
		db.ID,
		db.AccountID,
		db.StripeSubscriptionID,
		db.StripeCustomerID,
		db.StripePriceID,
		db.Status,
		fromNullTime(db.CurrentPeriodStart),
		fromNullTime(db.CurrentPeriodEnd),
		db.CancelAtPeriodEnd,
		fromNullTime(db.CanceledAt),
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	)
}

func toSubscriptionsDomain(dbs []subscriptionDB) []subscription.Subscription {
	subs := make([]subscription.Subscription, len(dbs))
	for i, db := range dbs {
		subs[i] = toSubscriptionDomain(db)
	}

	return subs
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
