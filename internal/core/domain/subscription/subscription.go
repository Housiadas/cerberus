package subscription

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("subscription not found")

type Subscription struct {
	id                 uuid.UUID
	accountID          uuid.UUID
	planID             uuid.UUID
	status             Status
	currentPeriodStart time.Time
	currentPeriodEnd   time.Time
	createdAt          time.Time
	updatedAt          time.Time
}

func New(
	id uuid.UUID,
	accountID uuid.UUID,
	planID uuid.UUID,
	status Status,
	currentPeriodStart time.Time,
	currentPeriodEnd time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) Subscription {
	return Subscription{
		id: id, accountID: accountID, planID: planID, status: status,
		currentPeriodStart: currentPeriodStart, currentPeriodEnd: currentPeriodEnd,
		createdAt: createdAt, updatedAt: updatedAt,
	}
}

func (s Subscription) ID() uuid.UUID                 { return s.id }
func (s Subscription) AccountID() uuid.UUID          { return s.accountID }
func (s Subscription) PlanID() uuid.UUID             { return s.planID }
func (s Subscription) Status() Status                { return s.status }
func (s Subscription) CurrentPeriodStart() time.Time { return s.currentPeriodStart }
func (s Subscription) CurrentPeriodEnd() time.Time   { return s.currentPeriodEnd }
func (s Subscription) CreatedAt() time.Time          { return s.createdAt }
func (s Subscription) UpdatedAt() time.Time          { return s.updatedAt }

func (s Subscription) WithStatus(st Status) Subscription {
	s.status = st

	return s
}

func (s Subscription) WithCurrentPeriodStart(t time.Time) Subscription {
	s.currentPeriodStart = t

	return s
}

func (s Subscription) WithCurrentPeriodEnd(t time.Time) Subscription {
	s.currentPeriodEnd = t

	return s
}

func (s Subscription) WithUpdatedAt(t time.Time) Subscription {
	s.updatedAt = t

	return s
}

type NewSubscription struct {
	AccountID          uuid.UUID
	PlanID             uuid.UUID
	Status             Status
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
}

type UpdateSubscription struct {
	Status             *Status
	CurrentPeriodStart *time.Time
	CurrentPeriodEnd   *time.Time
}

type QueryFilter struct {
	ID        *uuid.UUID
	AccountID *uuid.UUID
	PlanID    *uuid.UUID
	Status    *Status
}
