package invoice

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("invoice not found")

// Invoice represents a local cache of a Stripe invoice.
type Invoice struct {
	id               uuid.UUID
	accountID        uuid.UUID
	stripeInvoiceID  string
	stripeCustomerID string
	status           string
	amountDue        int64
	amountPaid       int64
	currency         string
	invoiceURL       string
	periodStart      *time.Time
	periodEnd        *time.Time
	createdAt        time.Time
	updatedAt        time.Time
}

// New constructs an Invoice from its individual fields.
func New(
	id uuid.UUID,
	accountID uuid.UUID,
	stripeInvoiceID string,
	stripeCustomerID string,
	status string,
	amountDue int64,
	amountPaid int64,
	currency string,
	invoiceURL string,
	periodStart *time.Time,
	periodEnd *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) Invoice {
	return Invoice{
		id:               id,
		accountID:        accountID,
		stripeInvoiceID:  stripeInvoiceID,
		stripeCustomerID: stripeCustomerID,
		status:           status,
		amountDue:        amountDue,
		amountPaid:       amountPaid,
		currency:         currency,
		invoiceURL:       invoiceURL,
		periodStart:      periodStart,
		periodEnd:        periodEnd,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
	}
}

// ID returns the invoice ID.
func (i Invoice) ID() uuid.UUID { return i.id }

// AccountID returns the account ID.
func (i Invoice) AccountID() uuid.UUID { return i.accountID }

// StripeInvoiceID returns the Stripe invoice ID.
func (i Invoice) StripeInvoiceID() string { return i.stripeInvoiceID }

// StripeCustomerID returns the Stripe customer ID.
func (i Invoice) StripeCustomerID() string { return i.stripeCustomerID }

// Status returns the invoice status.
func (i Invoice) Status() string { return i.status }

// AmountDue returns the amount due.
func (i Invoice) AmountDue() int64 { return i.amountDue }

// AmountPaid returns the amount paid.
func (i Invoice) AmountPaid() int64 { return i.amountPaid }

// Currency returns the currency.
func (i Invoice) Currency() string { return i.currency }

// InvoiceURL returns the invoice URL.
func (i Invoice) InvoiceURL() string { return i.invoiceURL }

// PeriodStart returns the period start.
func (i Invoice) PeriodStart() *time.Time { return i.periodStart }

// PeriodEnd returns the period end.
func (i Invoice) PeriodEnd() *time.Time { return i.periodEnd }

// CreatedAt returns the creation time.
func (i Invoice) CreatedAt() time.Time { return i.createdAt }

// UpdatedAt returns the last update time.
func (i Invoice) UpdatedAt() time.Time { return i.updatedAt }

// WithStatus returns a new Invoice with the given status.
func (i Invoice) WithStatus(status string) Invoice {
	i.status = status

	return i
}

// WithAmountPaid returns a new Invoice with the given amount paid.
func (i Invoice) WithAmountPaid(amount int64) Invoice {
	i.amountPaid = amount

	return i
}

// WithUpdatedAt returns a new Invoice with the given updated time.
func (i Invoice) WithUpdatedAt(t time.Time) Invoice {
	i.updatedAt = t

	return i
}
