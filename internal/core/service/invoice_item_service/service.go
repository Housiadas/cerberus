// Package invoice_item_service provides internal access to invoice item core.
package invoice_item_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/invoice_item"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for invoice item access.
type Service struct {
	log     logger.Logger
	storer  invoice_item.Storer
	uuidGen uuidgen.Generator
}

// New constructor.
func New(log logger.Logger, storer invoice_item.Storer, uuidGen uuidgen.Generator) *Service {
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
		return nil, fmt.Errorf("invoice item transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new invoice_item.InvoiceItem to the system.
func (c *Service) Create(
	ctx context.Context,
	nii invoice_item.NewInvoiceItem,
) (invoice_item.InvoiceItem, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return invoice_item.InvoiceItem{}, fmt.Errorf("invoice item uuid generate: %w", err)
	}

	now := time.Now()
	item := invoice_item.New(
		id,
		nii.InvoiceID,
		nii.Description,
		nii.Quantity,
		nii.UnitPriceCents,
		nii.TaxRateID,
		0,
		now,
	)

	err = c.storer.Create(ctx, item)
	if err != nil {
		return invoice_item.InvoiceItem{}, fmt.Errorf("invoice item create: %w", err)
	}

	return item, nil
}

// QueryByID finds the invoice item by the specified ID.
func (c *Service) QueryByID(
	ctx context.Context,
	invoiceItemID uuid.UUID,
) (invoice_item.InvoiceItem, error) {
	item, err := c.storer.QueryByID(ctx, invoiceItemID)
	if err != nil {
		return invoice_item.InvoiceItem{}, fmt.Errorf(
			"query: invoiceItemID[%s]: %w",
			invoiceItemID,
			err,
		)
	}

	return item, nil
}

// Query retrieves a list of existing invoice items.
func (c *Service) Query(
	ctx context.Context,
	filter invoice_item.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]invoice_item.InvoiceItem, error) {
	items, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("invoice item query: %w", err)
	}

	return items, nil
}

// QueryByInvoiceID retrieves invoice items for a specific invoice.
func (c *Service) QueryByInvoiceID(
	ctx context.Context,
	invoiceID uuid.UUID,
) ([]invoice_item.InvoiceItem, error) {
	items, err := c.storer.QueryByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("query: invoiceID[%s]: %w", invoiceID, err)
	}

	return items, nil
}
