package refund

import (
	"errors"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/payment"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("refund not found")

type Refund struct {
	id          uuid.UUID
	paymentID   uuid.UUID
	amountCents int
	reason      string
	status      payment.Status
	createdAt   time.Time
	updatedAt   time.Time
}

func New(
	id uuid.UUID,
	paymentID uuid.UUID,
	amountCents int,
	reason string,
	status payment.Status,
	createdAt time.Time,
	updatedAt time.Time,
) Refund {
	return Refund{
		id: id, paymentID: paymentID, amountCents: amountCents,
		reason: reason, status: status, createdAt: createdAt, updatedAt: updatedAt,
	}
}

func (r Refund) ID() uuid.UUID          { return r.id }
func (r Refund) PaymentID() uuid.UUID   { return r.paymentID }
func (r Refund) AmountCents() int       { return r.amountCents }
func (r Refund) Reason() string         { return r.reason }
func (r Refund) Status() payment.Status { return r.status }
func (r Refund) CreatedAt() time.Time   { return r.createdAt }
func (r Refund) UpdatedAt() time.Time   { return r.updatedAt }

func (r Refund) WithStatus(s payment.Status) Refund {
	r.status = s

	return r
}

func (r Refund) WithUpdatedAt(t time.Time) Refund {
	r.updatedAt = t

	return r
}

type NewRefund struct {
	PaymentID   uuid.UUID
	AmountCents int
	Reason      string
}

type QueryFilter struct {
	ID        *uuid.UUID
	PaymentID *uuid.UUID
	Status    *payment.Status
}
