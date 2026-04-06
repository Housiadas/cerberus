package payment_repo

import (
	"database/sql"
	"time"

	"github.com/Housiadas/cerberus/internal/core/payment"
	"github.com/google/uuid"
)

type paymentDB struct {
	ID               uuid.UUID      `db:"id"`
	AccountID        uuid.UUID      `db:"account_id"`
	StripePaymentID  string         `db:"stripe_payment_id"`
	StripeInvoiceID  sql.NullString `db:"stripe_invoice_id"`
	StripeCustomerID string         `db:"stripe_customer_id"`
	Amount           int64          `db:"amount"`
	Currency         string         `db:"currency"`
	Status           string         `db:"status"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
}

func toPaymentDB(pmt payment.Payment) paymentDB {
	return paymentDB{
		ID:               pmt.ID(),
		AccountID:        pmt.AccountID(),
		StripePaymentID:  pmt.StripePaymentID(),
		StripeInvoiceID:  toNullString(pmt.StripeInvoiceID()),
		StripeCustomerID: pmt.StripeCustomerID(),
		Amount:           pmt.Amount(),
		Currency:         pmt.Currency(),
		Status:           pmt.Status(),
		CreatedAt:        pmt.CreatedAt().UTC(),
		UpdatedAt:        pmt.UpdatedAt().UTC(),
	}
}

func toPaymentDomain(db paymentDB) payment.Payment {
	return payment.New(
		db.ID,
		db.AccountID,
		db.StripePaymentID,
		db.StripeInvoiceID.String,
		db.StripeCustomerID,
		db.Amount,
		db.Currency,
		db.Status,
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	)
}

func toPaymentsDomain(dbs []paymentDB) []payment.Payment {
	pmts := make([]payment.Payment, len(dbs))
	for i, db := range dbs {
		pmts[i] = toPaymentDomain(db)
	}

	return pmts
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}

	return sql.NullString{String: s, Valid: true}
}
