// Package tax_rate_service provides internal access to tax rate core.
package tax_rate_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/tax_rate"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for tax rate access.
type Service struct {
	log     logger.Logger
	storer  tax_rate.Storer
	uuidGen uuidgen.Generator
}

// New constructor.
func New(log logger.Logger, storer tax_rate.Storer, uuidGen uuidgen.Generator) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
	}
}

// NewWithTx constructs a new internal value that will use the
// specified transaction in any store-related calls.
func (c *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := c.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("tax rate transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new tax_rate.TaxRate to the system.
func (c *Service) Create(ctx context.Context, nt tax_rate.NewTaxRate) (tax_rate.TaxRate, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return tax_rate.TaxRate{}, fmt.Errorf("tax rate uuid generate: %w", err)
	}

	now := time.Now()
	tr := tax_rate.New(id, nt.Name, nt.Percentage, nt.Country, nt.Description, true, now, now)

	err = c.storer.Create(ctx, tr)
	if err != nil {
		return tax_rate.TaxRate{}, fmt.Errorf("tax rate create: %w", err)
	}

	return tr, nil
}

// Update modifies information about a tax_rate.TaxRate.
func (c *Service) Update(
	ctx context.Context,
	tr tax_rate.TaxRate,
	ut tax_rate.UpdateTaxRate,
) (tax_rate.TaxRate, error) {
	if ut.Name != nil {
		tr = tr.WithName(*ut.Name)
	}
	if ut.Percentage != nil {
		tr = tr.WithPercentage(*ut.Percentage)
	}
	if ut.Country != nil {
		tr = tr.WithCountry(*ut.Country)
	}
	if ut.Description != nil {
		tr = tr.WithDescription(*ut.Description)
	}
	if ut.IsActive != nil {
		tr = tr.WithIsActive(*ut.IsActive)
	}

	tr = tr.WithUpdatedAt(time.Now())

	err := c.storer.Update(ctx, tr)
	if err != nil {
		return tax_rate.TaxRate{}, fmt.Errorf("tax rate update: %w", err)
	}

	return tr, nil
}

// Delete removes the specified tax_rate.TaxRate.
func (c *Service) Delete(ctx context.Context, tr tax_rate.TaxRate) error {
	err := c.storer.Delete(ctx, tr)
	if err != nil {
		return fmt.Errorf("tax rate delete: %w", err)
	}

	return nil
}

// QueryByID finds the tax rate by the specified ID.
func (c *Service) QueryByID(ctx context.Context, taxRateID uuid.UUID) (tax_rate.TaxRate, error) {
	tr, err := c.storer.QueryByID(ctx, taxRateID)
	if err != nil {
		return tax_rate.TaxRate{}, fmt.Errorf("query: taxRateID[%s]: %w", taxRateID, err)
	}

	return tr, nil
}

// Query retrieves a list of existing tax rates.
func (c *Service) Query(
	ctx context.Context,
	filter tax_rate.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]tax_rate.TaxRate, error) {
	taxRates, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("tax rate query: %w", err)
	}

	return taxRates, nil
}
