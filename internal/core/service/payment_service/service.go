// Package payment_service provides internal access to payment core.
package payment_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/payment"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for payment access.
type Service struct {
	log     logger.Logger
	storer  payment.Storer
	uuidGen uuidgen.Generator
}

// New constructor.
func New(log logger.Logger, storer payment.Storer, uuidGen uuidgen.Generator) *Service {
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
		return nil, fmt.Errorf("payment transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new payment.Payment to the system.
func (c *Service) Create(ctx context.Context, np payment.NewPayment) (payment.Payment, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return payment.Payment{}, fmt.Errorf("payment uuid generate: %w", err)
	}

	now := time.Now()
	pmt := payment.New(
		id,
		np.InvoiceID,
		np.PaymentMethodID,
		np.AmountCents,
		np.Currency,
		np.Status,
		nil,
		now,
		now,
	)

	err = c.storer.Create(ctx, pmt)
	if err != nil {
		return payment.Payment{}, fmt.Errorf("payment create: %w", err)
	}

	return pmt, nil
}

// Update modifies information about a payment.Payment.
func (c *Service) Update(
	ctx context.Context,
	pmt payment.Payment,
	up payment.UpdatePayment,
) (payment.Payment, error) {
	if up.Status != nil {
		pmt = pmt.WithStatus(*up.Status)
	}

	if up.PaidAt != nil {
		pmt = pmt.WithPaidAt(up.PaidAt)
	}

	pmt = pmt.WithUpdatedAt(time.Now())

	err := c.storer.Update(ctx, pmt)
	if err != nil {
		return payment.Payment{}, fmt.Errorf("payment update: %w", err)
	}

	return pmt, nil
}

// QueryByID finds the payment by the specified ID.
func (c *Service) QueryByID(ctx context.Context, paymentID uuid.UUID) (payment.Payment, error) {
	pmt, err := c.storer.QueryByID(ctx, paymentID)
	if err != nil {
		return payment.Payment{}, fmt.Errorf("query: paymentID[%s]: %w", paymentID, err)
	}

	return pmt, nil
}

// QueryByInvoiceID finds payments by invoice ID.
func (c *Service) QueryByInvoiceID(
	ctx context.Context,
	invoiceID uuid.UUID,
) ([]payment.Payment, error) {
	pmts, err := c.storer.QueryByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("query: invoiceID[%s]: %w", invoiceID, err)
	}

	return pmts, nil
}

// Query retrieves a list of existing payments.
func (c *Service) Query(
	ctx context.Context,
	filter payment.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]payment.Payment, error) {
	payments, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("payment query: %w", err)
	}

	return payments, nil
}
