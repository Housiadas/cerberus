package invoice

import (
	"errors"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("invoice not found")

type Invoice struct {
	id             uuid.UUID
	accountID      uuid.UUID
	subscriptionID *uuid.UUID
	status         Status
	currency       currency.Currency
	dueDate        *time.Time
	issuedAt       *time.Time
	paidAt         *time.Time
	createdAt      time.Time
	updatedAt      time.Time
}

func New(
	id uuid.UUID,
	accountID uuid.UUID,
	subscriptionID *uuid.UUID,
	status Status,
	cur currency.Currency,
	dueDate *time.Time,
	issuedAt *time.Time,
	paidAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) Invoice {
	return Invoice{
		id: id, accountID: accountID, subscriptionID: subscriptionID,
		status: status, currency: cur, dueDate: dueDate, issuedAt: issuedAt,
		paidAt: paidAt, createdAt: createdAt, updatedAt: updatedAt,
	}
}

func (i Invoice) ID() uuid.UUID               { return i.id }
func (i Invoice) AccountID() uuid.UUID        { return i.accountID }
func (i Invoice) SubscriptionID() *uuid.UUID  { return i.subscriptionID }
func (i Invoice) Status() Status              { return i.status }
func (i Invoice) Currency() currency.Currency { return i.currency }
func (i Invoice) DueDate() *time.Time         { return i.dueDate }
func (i Invoice) IssuedAt() *time.Time        { return i.issuedAt }
func (i Invoice) PaidAt() *time.Time          { return i.paidAt }
func (i Invoice) CreatedAt() time.Time        { return i.createdAt }
func (i Invoice) UpdatedAt() time.Time        { return i.updatedAt }

func (i Invoice) WithStatus(s Status) Invoice {
	i.status = s

	return i
}

func (i Invoice) WithDueDate(t *time.Time) Invoice {
	i.dueDate = t

	return i
}

func (i Invoice) WithIssuedAt(t *time.Time) Invoice {
	i.issuedAt = t

	return i
}

func (i Invoice) WithPaidAt(t *time.Time) Invoice {
	i.paidAt = t

	return i
}

func (i Invoice) WithUpdatedAt(t time.Time) Invoice {
	i.updatedAt = t

	return i
}

type NewInvoice struct {
	AccountID      uuid.UUID
	SubscriptionID *uuid.UUID
	Status         Status
	Currency       currency.Currency
	DueDate        *time.Time
}

type UpdateInvoice struct {
	Status   *Status
	DueDate  *time.Time
	IssuedAt *time.Time
	PaidAt   *time.Time
}

type QueryFilter struct {
	ID        *uuid.UUID
	AccountID *uuid.UUID
	Status    *Status
}
