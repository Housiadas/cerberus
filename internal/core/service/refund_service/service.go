// Package refund_service provides internal access to refund core.
package refund_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/payment"
	"github.com/Housiadas/cerberus/internal/core/domain/refund"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for refund access.
type Service struct {
	log     logger.Logger
	storer  refund.Storer
	uuidGen uuidgen.Generator
}

// New constructor.
func New(log logger.Logger, storer refund.Storer, uuidGen uuidgen.Generator) *Service {
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
		return nil, fmt.Errorf("refund transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new refund.Refund to the system.
func (c *Service) Create(ctx context.Context, nr refund.NewRefund) (refund.Refund, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return refund.Refund{}, fmt.Errorf("refund uuid generate: %w", err)
	}

	now := time.Now()
	ref := refund.New(id, nr.PaymentID, nr.AmountCents, nr.Reason, payment.MustParse("pending"), now, now)

	err = c.storer.Create(ctx, ref)
	if err != nil {
		return refund.Refund{}, fmt.Errorf("refund create: %w", err)
	}

	return ref, nil
}

// QueryByID finds the refund by the specified ID.
func (c *Service) QueryByID(ctx context.Context, refundID uuid.UUID) (refund.Refund, error) {
	ref, err := c.storer.QueryByID(ctx, refundID)
	if err != nil {
		return refund.Refund{}, fmt.Errorf("query: refundID[%s]: %w", refundID, err)
	}

	return ref, nil
}

// Query retrieves a list of existing refunds.
func (c *Service) Query(
	ctx context.Context,
	filter refund.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]refund.Refund, error) {
	refs, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("refund query: %w", err)
	}

	return refs, nil
}

// QueryByPaymentID retrieves refunds for a specific payment.
func (c *Service) QueryByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]refund.Refund, error) {
	refs, err := c.storer.QueryByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("query: paymentID[%s]: %w", paymentID, err)
	}

	return refs, nil
}
