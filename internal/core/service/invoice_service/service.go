// Package invoice_service provides internal access to invoice core.
package invoice_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/invoice"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for invoice access.
type Service struct {
	log     logger.Logger
	storer  invoice.Storer
	uuidGen uuidgen.Generator
}

// New constructor.
func New(log logger.Logger, storer invoice.Storer, uuidGen uuidgen.Generator) *Service {
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
		return nil, fmt.Errorf("invoice transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new invoice.Invoice to the system.
func (c *Service) Create(ctx context.Context, ni invoice.NewInvoice) (invoice.Invoice, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("invoice uuid generate: %w", err)
	}

	now := time.Now()
	inv := invoice.New(id, ni.AccountID, ni.SubscriptionID, ni.Status, ni.Currency, ni.DueDate, nil, nil, now, now)

	err = c.storer.Create(ctx, inv)
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("invoice create: %w", err)
	}

	return inv, nil
}

// Update modifies information about an invoice.Invoice.
func (c *Service) Update(
	ctx context.Context,
	inv invoice.Invoice,
	ui invoice.UpdateInvoice,
) (invoice.Invoice, error) {
	if ui.Status != nil {
		inv = inv.WithStatus(*ui.Status)
	}
	if ui.DueDate != nil {
		inv = inv.WithDueDate(ui.DueDate)
	}
	if ui.IssuedAt != nil {
		inv = inv.WithIssuedAt(ui.IssuedAt)
	}
	if ui.PaidAt != nil {
		inv = inv.WithPaidAt(ui.PaidAt)
	}

	inv = inv.WithUpdatedAt(time.Now())

	err := c.storer.Update(ctx, inv)
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("invoice update: %w", err)
	}

	return inv, nil
}

// QueryByID finds the invoice by the specified ID.
func (c *Service) QueryByID(ctx context.Context, invoiceID uuid.UUID) (invoice.Invoice, error) {
	inv, err := c.storer.QueryByID(ctx, invoiceID)
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("query: invoiceID[%s]: %w", invoiceID, err)
	}

	return inv, nil
}

// QueryByAccountID finds invoices by account ID.
func (c *Service) QueryByAccountID(ctx context.Context, accountID uuid.UUID) ([]invoice.Invoice, error) {
	invs, err := c.storer.QueryByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("query: accountID[%s]: %w", accountID, err)
	}

	return invs, nil
}

// Query retrieves a list of existing invoices.
func (c *Service) Query(
	ctx context.Context,
	filter invoice.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]invoice.Invoice, error) {
	invoices, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("invoice query: %w", err)
	}

	return invoices, nil
}
