package refund

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("refund not found")

// Refund represents a refund record.
type Refund struct {
	id             uuid.UUID
	accountID      uuid.UUID
	paymentID      uuid.UUID
	stripeRefundID string
	amount         int64
	currency       string
	status         string
	reason         string
	createdAt      time.Time
	updatedAt      time.Time
}

// New constructs a Refund from its individual fields.
func New(
	id uuid.UUID,
	accountID uuid.UUID,
	paymentID uuid.UUID,
	stripeRefundID string,
	amount int64,
	currency string,
	status string,
	reason string,
	createdAt time.Time,
	updatedAt time.Time,
) Refund {
	return Refund{
		id:             id,
		accountID:      accountID,
		paymentID:      paymentID,
		stripeRefundID: stripeRefundID,
		amount:         amount,
		currency:       currency,
		status:         status,
		reason:         reason,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

// ID returns the refund ID.
func (r Refund) ID() uuid.UUID { return r.id }

// AccountID returns the account ID.
func (r Refund) AccountID() uuid.UUID { return r.accountID }

// PaymentID returns the payment ID.
func (r Refund) PaymentID() uuid.UUID { return r.paymentID }

// StripeRefundID returns the Stripe refund ID.
func (r Refund) StripeRefundID() string { return r.stripeRefundID }

// Amount returns the refund amount.
func (r Refund) Amount() int64 { return r.amount }

// Currency returns the currency.
func (r Refund) Currency() string { return r.currency }

// Status returns the refund status.
func (r Refund) Status() string { return r.status }

// Reason returns the refund reason.
func (r Refund) Reason() string { return r.reason }

// CreatedAt returns the creation time.
func (r Refund) CreatedAt() time.Time { return r.createdAt }

// UpdatedAt returns the last update time.
func (r Refund) UpdatedAt() time.Time { return r.updatedAt }
