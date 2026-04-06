package subscription

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("subscription not found")

// Subscription represents a local cache of a Stripe subscription.
type Subscription struct {
	id                   uuid.UUID
	accountID            uuid.UUID
	stripeSubscriptionID string
	stripeCustomerID     string
	stripePriceID        string
	status               string
	currentPeriodStart   *time.Time
	currentPeriodEnd     *time.Time
	cancelAtPeriodEnd    bool
	canceledAt           *time.Time
	createdAt            time.Time
	updatedAt            time.Time
}

// New constructs a Subscription from its individual fields.
func New(
	id uuid.UUID,
	accountID uuid.UUID,
	stripeSubscriptionID string,
	stripeCustomerID string,
	stripePriceID string,
	status string,
	currentPeriodStart *time.Time,
	currentPeriodEnd *time.Time,
	cancelAtPeriodEnd bool,
	canceledAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) Subscription {
	return Subscription{
		id:                   id,
		accountID:            accountID,
		stripeSubscriptionID: stripeSubscriptionID,
		stripeCustomerID:     stripeCustomerID,
		stripePriceID:        stripePriceID,
		status:               status,
		currentPeriodStart:   currentPeriodStart,
		currentPeriodEnd:     currentPeriodEnd,
		cancelAtPeriodEnd:    cancelAtPeriodEnd,
		canceledAt:           canceledAt,
		createdAt:            createdAt,
		updatedAt:            updatedAt,
	}
}

// ID returns the subscription ID.
func (s Subscription) ID() uuid.UUID { return s.id }

// AccountID returns the account ID.
func (s Subscription) AccountID() uuid.UUID { return s.accountID }

// StripeSubscriptionID returns the Stripe subscription ID.
func (s Subscription) StripeSubscriptionID() string { return s.stripeSubscriptionID }

// StripeCustomerID returns the Stripe customer ID.
func (s Subscription) StripeCustomerID() string { return s.stripeCustomerID }

// StripePriceID returns the Stripe price ID.
func (s Subscription) StripePriceID() string { return s.stripePriceID }

// Status returns the subscription status.
func (s Subscription) Status() string { return s.status }

// CurrentPeriodStart returns the current period start time.
func (s Subscription) CurrentPeriodStart() *time.Time { return s.currentPeriodStart }

// CurrentPeriodEnd returns the current period end time.
func (s Subscription) CurrentPeriodEnd() *time.Time { return s.currentPeriodEnd }

// CancelAtPeriodEnd returns whether the subscription cancels at period end.
func (s Subscription) CancelAtPeriodEnd() bool { return s.cancelAtPeriodEnd }

// CanceledAt returns the cancellation time.
func (s Subscription) CanceledAt() *time.Time { return s.canceledAt }

// CreatedAt returns the creation time.
func (s Subscription) CreatedAt() time.Time { return s.createdAt }

// UpdatedAt returns the last update time.
func (s Subscription) UpdatedAt() time.Time { return s.updatedAt }

// WithStatus returns a new Subscription with the given status.
func (s Subscription) WithStatus(status string) Subscription {
	s.status = status

	return s
}

// WithCurrentPeriodStart returns a new Subscription with the given period start.
func (s Subscription) WithCurrentPeriodStart(t *time.Time) Subscription {
	s.currentPeriodStart = t

	return s
}

// WithCurrentPeriodEnd returns a new Subscription with the given period end.
func (s Subscription) WithCurrentPeriodEnd(t *time.Time) Subscription {
	s.currentPeriodEnd = t

	return s
}

// WithCancelAtPeriodEnd returns a new Subscription with the given cancel flag.
func (s Subscription) WithCancelAtPeriodEnd(cancel bool) Subscription {
	s.cancelAtPeriodEnd = cancel

	return s
}

// WithCanceledAt returns a new Subscription with the given canceled time.
func (s Subscription) WithCanceledAt(t *time.Time) Subscription {
	s.canceledAt = t

	return s
}

// WithUpdatedAt returns a new Subscription with the given updated time.
func (s Subscription) WithUpdatedAt(t time.Time) Subscription {
	s.updatedAt = t

	return s
}
