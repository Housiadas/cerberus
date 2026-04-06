package invoice_repo

import (
	"database/sql"
	"time"

	"github.com/Housiadas/cerberus/internal/core/invoice"
	"github.com/google/uuid"
)

type invoiceDB struct {
	ID               uuid.UUID      `db:"id"`
	AccountID        uuid.UUID      `db:"account_id"`
	StripeInvoiceID  string         `db:"stripe_invoice_id"`
	StripeCustomerID string         `db:"stripe_customer_id"`
	Status           string         `db:"status"`
	AmountDue        int64          `db:"amount_due"`
	AmountPaid       int64          `db:"amount_paid"`
	Currency         string         `db:"currency"`
	InvoiceURL       sql.NullString `db:"invoice_url"`
	PeriodStart      sql.NullTime   `db:"period_start"`
	PeriodEnd        sql.NullTime   `db:"period_end"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
}

func toInvoiceDB(inv invoice.Invoice) invoiceDB {
	return invoiceDB{
		ID:               inv.ID(),
		AccountID:        inv.AccountID(),
		StripeInvoiceID:  inv.StripeInvoiceID(),
		StripeCustomerID: inv.StripeCustomerID(),
		Status:           inv.Status(),
		AmountDue:        inv.AmountDue(),
		AmountPaid:       inv.AmountPaid(),
		Currency:         inv.Currency(),
		InvoiceURL:       toNullString(inv.InvoiceURL()),
		PeriodStart:      toNullTime(inv.PeriodStart()),
		PeriodEnd:        toNullTime(inv.PeriodEnd()),
		CreatedAt:        inv.CreatedAt().UTC(),
		UpdatedAt:        inv.UpdatedAt().UTC(),
	}
}

func toInvoiceDomain(db invoiceDB) invoice.Invoice {
	return invoice.New(
		db.ID,
		db.AccountID,
		db.StripeInvoiceID,
		db.StripeCustomerID,
		db.Status,
		db.AmountDue,
		db.AmountPaid,
		db.Currency,
		db.InvoiceURL.String,
		fromNullTime(db.PeriodStart),
		fromNullTime(db.PeriodEnd),
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	)
}

func toInvoicesDomain(dbs []invoiceDB) []invoice.Invoice {
	invs := make([]invoice.Invoice, len(dbs))
	for i, db := range dbs {
		invs[i] = toInvoiceDomain(db)
	}

	return invs
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: t.UTC(), Valid: true}
}

func fromNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}

	t := nt.Time.In(time.UTC)

	return &t
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}

	return sql.NullString{String: s, Valid: true}
}
