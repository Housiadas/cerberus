package invoice_item_repo

import (
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/invoice_item"
	"github.com/google/uuid"
)

type invoiceItemDB struct {
	ID             uuid.UUID     `db:"id"`
	InvoiceID      uuid.UUID     `db:"invoice_id"`
	Description    string        `db:"description"`
	Quantity       int           `db:"quantity"`
	UnitPriceCents int           `db:"unit_price_cents"`
	TaxRateID      uuid.NullUUID `db:"tax_rate_id"`
	TotalCents     int           `db:"total_cents"`
	CreatedAt      time.Time     `db:"created_at"`
}

func toInvoiceItemDB(item invoice_item.InvoiceItem) invoiceItemDB {
	return invoiceItemDB{
		ID:             item.ID(),
		InvoiceID:      item.InvoiceID(),
		Description:    item.Description(),
		Quantity:       item.Quantity(),
		UnitPriceCents: item.UnitPriceCents(),
		TaxRateID:      toNullUUID(item.TaxRateID()),
		CreatedAt:      item.CreatedAt().UTC(),
	}
}

func toDomain(db invoiceItemDB) invoice_item.InvoiceItem {
	return invoice_item.New(
		db.ID,
		db.InvoiceID,
		db.Description,
		db.Quantity,
		db.UnitPriceCents,
		fromNullUUID(db.TaxRateID),
		db.TotalCents,
		db.CreatedAt.In(time.UTC),
	)
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

func toSliceDomain(dbs []invoiceItemDB) []invoice_item.InvoiceItem {
	items := make([]invoice_item.InvoiceItem, len(dbs))

	for i, db := range dbs {
		items[i] = toDomain(db)
	}

	return items
}
