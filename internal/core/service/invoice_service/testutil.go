package invoice_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/Housiadas/cerberus/internal/core/domain/invoice"
	"github.com/google/uuid"
)

// TestNewInvoices is a helper method for testing.
func TestNewInvoices(n int) []invoice.NewInvoice {
	newInvoices := make([]invoice.NewInvoice, n)

	for i := range n {
		ni := invoice.NewInvoice{
			AccountID:      uuid.New(),
			SubscriptionID: nil,
			Status:         invoice.MustParse("draft"),
			Currency:       currency.MustParse("USD"),
			DueDate:        nil,
		}

		newInvoices[i] = ni
	}

	return newInvoices
}

// TestSeedInvoices is a helper method for testing.
func TestSeedInvoices(ctx context.Context, n int, service *Service) ([]invoice.Invoice, error) {
	newInvoices := TestNewInvoices(n)

	invoices := make([]invoice.Invoice, len(newInvoices))

	for i, ni := range newInvoices {
		inv, err := service.Create(ctx, ni)
		if err != nil {
			return nil, fmt.Errorf("seeding invoice: idx: %d : %w", i, err)
		}

		invoices[i] = inv
	}

	return invoices, nil
}
