package tax_rate_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/country"
	"github.com/Housiadas/cerberus/internal/core/domain/tax_rate"
)

// TestNewTaxRates is a helper method for testing.
func TestNewTaxRates(n int) []tax_rate.NewTaxRate {
	newTaxRates := make([]tax_rate.NewTaxRate, n)

	for i := range n {
		nt := tax_rate.NewTaxRate{
			Name:        fmt.Sprintf("TaxRate%d", i),
			Percentage:  20.0,
			Country:     country.MustParse("US"),
			Description: fmt.Sprintf("Description%d", i),
		}

		newTaxRates[i] = nt
	}

	return newTaxRates
}

// TestSeedTaxRates is a helper method for testing.
func TestSeedTaxRates(ctx context.Context, n int, service *Service) ([]tax_rate.TaxRate, error) {
	newTaxRates := TestNewTaxRates(n)

	taxRates := make([]tax_rate.TaxRate, len(newTaxRates))

	for i, nt := range newTaxRates {
		tr, err := service.Create(ctx, nt)
		if err != nil {
			return nil, fmt.Errorf("seeding tax rate: idx: %d : %w", i, err)
		}

		taxRates[i] = tr
	}

	return taxRates, nil
}
