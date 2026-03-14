package invoice_item

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("invoice item not found")

type InvoiceItem struct {
	id             uuid.UUID
	invoiceID      uuid.UUID
	description    string
	quantity       int
	unitPriceCents int
	taxRateID      *uuid.UUID
	totalCents     int
	createdAt      time.Time
}

func New(
	id uuid.UUID,
	invoiceID uuid.UUID,
	description string,
	quantity int,
	unitPriceCents int,
	taxRateID *uuid.UUID,
	totalCents int,
	createdAt time.Time,
) InvoiceItem {
	return InvoiceItem{
		id: id, invoiceID: invoiceID, description: description,
		quantity: quantity, unitPriceCents: unitPriceCents, taxRateID: taxRateID,
		totalCents: totalCents, createdAt: createdAt,
	}
}

func (i InvoiceItem) ID() uuid.UUID         { return i.id }
func (i InvoiceItem) InvoiceID() uuid.UUID  { return i.invoiceID }
func (i InvoiceItem) Description() string   { return i.description }
func (i InvoiceItem) Quantity() int         { return i.quantity }
func (i InvoiceItem) UnitPriceCents() int   { return i.unitPriceCents }
func (i InvoiceItem) TaxRateID() *uuid.UUID { return i.taxRateID }
func (i InvoiceItem) TotalCents() int       { return i.totalCents }
func (i InvoiceItem) CreatedAt() time.Time  { return i.createdAt }

type NewInvoiceItem struct {
	InvoiceID      uuid.UUID
	Description    string
	Quantity       int
	UnitPriceCents int
	TaxRateID      *uuid.UUID
}

type QueryFilter struct {
	ID        *uuid.UUID
	InvoiceID *uuid.UUID
}
