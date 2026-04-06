package payment

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("payment not found")

// Payment represents a local cache of a Stripe payment.
type Payment struct {
	id               uuid.UUID
	accountID        uuid.UUID
	stripePaymentID  string
	stripeInvoiceID  string
	stripeCustomerID string
	amount           int64
	currency         string
	status           string
	createdAt        time.Time
	updatedAt        time.Time
}

// New constructs a Payment from its individual fields.
func New(
	id uuid.UUID,
	accountID uuid.UUID,
	stripePaymentID string,
	stripeInvoiceID string,
	stripeCustomerID string,
	amount int64,
	currency string,
	status string,
	createdAt time.Time,
	updatedAt time.Time,
) Payment {
	return Payment{
		id:               id,
		accountID:        accountID,
		stripePaymentID:  stripePaymentID,
		stripeInvoiceID:  stripeInvoiceID,
		stripeCustomerID: stripeCustomerID,
		amount:           amount,
		currency:         currency,
		status:           status,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
	}
}

// ID returns the payment ID.
func (p Payment) ID() uuid.UUID { return p.id }

// AccountID returns the account ID.
func (p Payment) AccountID() uuid.UUID { return p.accountID }

// StripePaymentID returns the Stripe payment ID.
func (p Payment) StripePaymentID() string { return p.stripePaymentID }

// StripeInvoiceID returns the Stripe invoice ID.
func (p Payment) StripeInvoiceID() string { return p.stripeInvoiceID }

// StripeCustomerID returns the Stripe customer ID.
func (p Payment) StripeCustomerID() string { return p.stripeCustomerID }

// Amount returns the payment amount.
func (p Payment) Amount() int64 { return p.amount }

// Currency returns the currency.
func (p Payment) Currency() string { return p.currency }

// Status returns the payment status.
func (p Payment) Status() string { return p.status }

// CreatedAt returns the creation time.
func (p Payment) CreatedAt() time.Time { return p.createdAt }

// UpdatedAt returns the last update time.
func (p Payment) UpdatedAt() time.Time { return p.updatedAt }

// WithStatus returns a new Payment with the given status.
func (p Payment) WithStatus(status string) Payment {
	p.status = status

	return p
}

// WithUpdatedAt returns a new Payment with the given updated time.
func (p Payment) WithUpdatedAt(t time.Time) Payment {
	p.updatedAt = t

	return p
}
