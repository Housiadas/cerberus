package tax_rate_repo

import (
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/country"
	"github.com/Housiadas/cerberus/internal/core/domain/tax_rate"
	"github.com/google/uuid"
)

type taxRateDB struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Percentage  float64   `db:"percentage"`
	Country     string    `db:"country"`
	Description string    `db:"description"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func toTaxRateDB(tr tax_rate.TaxRate) taxRateDB {
	return taxRateDB{
		ID:          tr.ID(),
		Name:        tr.Name(),
		Percentage:  tr.Percentage(),
		Country:     tr.Country().String(),
		Description: tr.Description(),
		IsActive:    tr.IsActive(),
		CreatedAt:   tr.CreatedAt().UTC(),
		UpdatedAt:   tr.UpdatedAt().UTC(),
	}
}

func toTaxRateDomain(db taxRateDB) (tax_rate.TaxRate, error) {
	cntry, err := country.Parse(db.Country)
	if err != nil {
		return tax_rate.TaxRate{}, fmt.Errorf("parse country: %w", err)
	}

	return tax_rate.New(
		db.ID,
		db.Name,
		db.Percentage,
		cntry,
		db.Description,
		db.IsActive,
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	), nil
}

func toTaxRatesDomain(dbs []taxRateDB) ([]tax_rate.TaxRate, error) {
	items := make([]tax_rate.TaxRate, len(dbs))

	for i, db := range dbs {
		var err error

		items[i], err = toTaxRateDomain(db)
		if err != nil {
			return nil, fmt.Errorf("to tax rates domain error: %w", err)
		}
	}

	return items, nil
}
