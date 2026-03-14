package subscription_repo

import (
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/subscription"
	"github.com/google/uuid"
)

type subscriptionDB struct {
	ID                 uuid.UUID `db:"id"`
	AccountID          uuid.UUID `db:"account_id"`
	PlanID             uuid.UUID `db:"plan_id"`
	Status             string    `db:"status"`
	CurrentPeriodStart time.Time `db:"current_period_start"`
	CurrentPeriodEnd   time.Time `db:"current_period_end"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

func toSubscriptionDB(sub subscription.Subscription) subscriptionDB {
	return subscriptionDB{
		ID:                 sub.ID(),
		AccountID:          sub.AccountID(),
		PlanID:             sub.PlanID(),
		Status:             sub.Status().String(),
		CurrentPeriodStart: sub.CurrentPeriodStart().UTC(),
		CurrentPeriodEnd:   sub.CurrentPeriodEnd().UTC(),
		CreatedAt:          sub.CreatedAt().UTC(),
		UpdatedAt:          sub.UpdatedAt().UTC(),
	}
}

func toDomain(db subscriptionDB) (subscription.Subscription, error) {
	status, err := subscription.Parse(db.Status)
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("parse status: %w", err)
	}

	return subscription.New(
		db.ID,
		db.AccountID,
		db.PlanID,
		status,
		db.CurrentPeriodStart.In(time.UTC),
		db.CurrentPeriodEnd.In(time.UTC),
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	), nil
}

func toSliceDomain(dbs []subscriptionDB) ([]subscription.Subscription, error) {
	subs := make([]subscription.Subscription, len(dbs))

	for i, db := range dbs {
		var err error

		subs[i], err = toDomain(db)
		if err != nil {
			return nil, fmt.Errorf("to subscriptions domain error: %w", err)
		}
	}

	return subs, nil
}
