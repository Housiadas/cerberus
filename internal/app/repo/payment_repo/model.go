package payment_repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/Housiadas/cerberus/internal/core/domain/payment"
	"github.com/google/uuid"
)

type paymentDB struct {
	ID              uuid.UUID    `db:"id"`
	InvoiceID       uuid.UUID    `db:"invoice_id"`
	PaymentMethodID uuid.UUID    `db:"payment_method_id"`
	AmountCents     int          `db:"amount_cents"`
	Currency        string       `db:"currency"`
	Status          string       `db:"status"`
	PaidAt          sql.NullTime `db:"paid_at"`
	CreatedAt       time.Time    `db:"created_at"`
	UpdatedAt       time.Time    `db:"updated_at"`
}

func toPaymentDB(pmt payment.Payment) paymentDB {
	return paymentDB{
		ID:              pmt.ID(),
		InvoiceID:       pmt.InvoiceID(),
		PaymentMethodID: pmt.PaymentMethodID(),
		AmountCents:     pmt.AmountCents(),
		Currency:        pmt.Currency().String(),
		Status:          pmt.Status().String(),
		PaidAt:          toNullTime(pmt.PaidAt()),
		CreatedAt:       pmt.CreatedAt().UTC(),
		UpdatedAt:       pmt.UpdatedAt().UTC(),
	}
}

func toDomain(db paymentDB) (payment.Payment, error) {
	cur, err := currency.Parse(db.Currency)
	if err != nil {
		return payment.Payment{}, fmt.Errorf("parse currency: %w", err)
	}

	status, err := payment.Parse(db.Status)
	if err != nil {
		return payment.Payment{}, fmt.Errorf("parse status: %w", err)
	}

	return payment.New(
		db.ID,
		db.InvoiceID,
		db.PaymentMethodID,
		db.AmountCents,
		cur,
		status,
		fromNullTime(db.PaidAt),
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	), nil
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

func toSliceDomain(dbs []paymentDB) ([]payment.Payment, error) {
	pmts := make([]payment.Payment, len(dbs))

	for i, db := range dbs {
		var err error

		pmts[i], err = toDomain(db)
		if err != nil {
			return nil, fmt.Errorf("to payments domain error: %w", err)
		}
	}

	return pmts, nil
}
