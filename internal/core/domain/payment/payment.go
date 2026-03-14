package payment

import (
	"errors"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("payment not found")

type Payment struct {
	id              uuid.UUID
	invoiceID       uuid.UUID
	paymentMethodID uuid.UUID
	amountCents     int
	currency        currency.Currency
	status          Status
	paidAt          *time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

func New(
	id uuid.UUID,
	invoiceID uuid.UUID,
	paymentMethodID uuid.UUID,
	amountCents int,
	cur currency.Currency,
	status Status,
	paidAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) Payment {
	return Payment{
		id: id, invoiceID: invoiceID, paymentMethodID: paymentMethodID,
		amountCents: amountCents, currency: cur, status: status,
		paidAt: paidAt, createdAt: createdAt, updatedAt: updatedAt,
	}
}

func (p Payment) ID() uuid.UUID               { return p.id }
func (p Payment) InvoiceID() uuid.UUID        { return p.invoiceID }
func (p Payment) PaymentMethodID() uuid.UUID  { return p.paymentMethodID }
func (p Payment) AmountCents() int            { return p.amountCents }
func (p Payment) Currency() currency.Currency { return p.currency }
func (p Payment) Status() Status              { return p.status }
func (p Payment) PaidAt() *time.Time          { return p.paidAt }
func (p Payment) CreatedAt() time.Time        { return p.createdAt }
func (p Payment) UpdatedAt() time.Time        { return p.updatedAt }

func (p Payment) WithStatus(s Status) Payment {
	p.status = s

	return p
}

func (p Payment) WithPaidAt(t *time.Time) Payment {
	p.paidAt = t

	return p
}

func (p Payment) WithUpdatedAt(t time.Time) Payment {
	p.updatedAt = t

	return p
}

type NewPayment struct {
	InvoiceID       uuid.UUID
	PaymentMethodID uuid.UUID
	AmountCents     int
	Currency        currency.Currency
	Status          Status
}

type UpdatePayment struct {
	Status *Status
	PaidAt *time.Time
}

type QueryFilter struct {
	ID        *uuid.UUID
	InvoiceID *uuid.UUID
	Status    *Status
}
