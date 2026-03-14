package payment_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/currency"
	"github.com/Housiadas/cerberus/internal/core/domain/payment"
	"github.com/google/uuid"
)

// TestNewPayments is a helper method for testing.
func TestNewPayments(n int) []payment.NewPayment {
	newPayments := make([]payment.NewPayment, n)

	for i := range n {
		np := payment.NewPayment{
			InvoiceID:       uuid.New(),
			PaymentMethodID: uuid.New(),
			AmountCents:     1000,
			Currency:        currency.MustParse("USD"),
			Status:          payment.MustParse("pending"),
		}

		newPayments[i] = np
	}

	return newPayments
}

// TestSeedPayments is a helper method for testing.
func TestSeedPayments(ctx context.Context, n int, service *Service) ([]payment.Payment, error) {
	newPayments := TestNewPayments(n)

	payments := make([]payment.Payment, len(newPayments))

	for i, np := range newPayments {
		pmt, err := service.Create(ctx, np)
		if err != nil {
			return nil, fmt.Errorf("seeding payment: idx: %d : %w", i, err)
		}

		payments[i] = pmt
	}

	return payments, nil
}
