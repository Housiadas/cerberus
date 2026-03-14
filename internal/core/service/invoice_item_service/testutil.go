package invoice_item_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/invoice_item"
	"github.com/google/uuid"
)

// TestNewInvoiceItems is a helper method for testing.
func TestNewInvoiceItems(n int) []invoice_item.NewInvoiceItem {
	newItems := make([]invoice_item.NewInvoiceItem, n)

	for i := range n {
		nitem := invoice_item.NewInvoiceItem{
			InvoiceID:      uuid.New(),
			Description:    fmt.Sprintf("Item%d", i),
			Quantity:       i + 1,
			UnitPriceCents: 1000 * (i + 1),
			TaxRateID:      nil,
		}

		newItems[i] = nitem
	}

	return newItems
}

// TestSeedInvoiceItems is a helper method for testing.
func TestSeedInvoiceItems(ctx context.Context, n int, service *Service) ([]invoice_item.InvoiceItem, error) {
	newItems := TestNewInvoiceItems(n)

	items := make([]invoice_item.InvoiceItem, len(newItems))

	for i, nitem := range newItems {
		item, err := service.Create(ctx, nitem)
		if err != nil {
			return nil, fmt.Errorf("seeding invoice item: idx: %d : %w", i, err)
		}

		items[i] = item
	}

	return items, nil
}
