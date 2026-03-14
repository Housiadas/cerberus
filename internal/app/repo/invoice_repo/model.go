package invoice_repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/Housiadas/cerberus/internal/core/domain/invoice"
	"github.com/google/uuid"
)

type invoiceDB struct {
	ID             uuid.UUID     `db:"id"`
	AccountID      uuid.UUID     `db:"account_id"`
	SubscriptionID uuid.NullUUID `db:"subscription_id"`
	Status         string        `db:"status"`
	Currency       string        `db:"currency"`
	DueDate        sql.NullTime  `db:"due_date"`
	IssuedAt       sql.NullTime  `db:"issued_at"`
	PaidAt         sql.NullTime  `db:"paid_at"`
	CreatedAt      time.Time     `db:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at"`
}

func toInvoiceDB(inv invoice.Invoice) invoiceDB {
	return invoiceDB{
		ID:             inv.ID(),
		AccountID:      inv.AccountID(),
		SubscriptionID: toNullUUID(inv.SubscriptionID()),
		Status:         inv.Status().String(),
		Currency:       inv.Currency().String(),
		DueDate:        toNullTime(inv.DueDate()),
		IssuedAt:       toNullTime(inv.IssuedAt()),
		PaidAt:         toNullTime(inv.PaidAt()),
		CreatedAt:      inv.CreatedAt().UTC(),
		UpdatedAt:      inv.UpdatedAt().UTC(),
	}
}

func toDomain(db invoiceDB) (invoice.Invoice, error) {
	status, err := invoice.Parse(db.Status)
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("parse status: %w", err)
	}

	cur, err := currency.Parse(db.Currency)
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("parse currency: %w", err)
	}

	return invoice.New(
		db.ID,
		db.AccountID,
		fromNullUUID(db.SubscriptionID),
		status,
		cur,
		fromNullTime(db.DueDate),
		fromNullTime(db.IssuedAt),
		fromNullTime(db.PaidAt),
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	), nil
}

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}

	return uuid.NullUUID{UUID: *id, Valid: true}
}

func fromNullUUID(nu uuid.NullUUID) *uuid.UUID {
	if !nu.Valid {
		return nil
	}

	return &nu.UUID
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

func toSliceDomain(dbs []invoiceDB) ([]invoice.Invoice, error) {
	invs := make([]invoice.Invoice, len(dbs))

	for i, db := range dbs {
		var err error

		invs[i], err = toDomain(db)
		if err != nil {
			return nil, fmt.Errorf("to invoices domain error: %w", err)
		}
	}

	return invs, nil
}
