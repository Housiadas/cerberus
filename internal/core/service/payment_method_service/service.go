// Package payment_method_service provides internal access to payment method core.
package payment_method_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/payment_method"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for payment method access.
type Service struct {
	log     logger.Logger
	storer  payment_method.Storer
	uuidGen uuidgen.Generator
}

// New constructor.
func New(log logger.Logger, storer payment_method.Storer, uuidGen uuidgen.Generator) *Service {
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
		return nil, fmt.Errorf("payment method transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new payment_method.PaymentMethod to the system.
func (c *Service) Create(ctx context.Context, npm payment_method.NewPaymentMethod) (payment_method.PaymentMethod, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return payment_method.PaymentMethod{}, fmt.Errorf("payment method uuid generate: %w", err)
	}

	now := time.Now()
	pm := payment_method.New(id, npm.AccountID, npm.Type, npm.Details, npm.IsDefault, now, now)

	err = c.storer.Create(ctx, pm)
	if err != nil {
		return payment_method.PaymentMethod{}, fmt.Errorf("payment method create: %w", err)
	}

	return pm, nil
}

// Delete removes the specified payment_method.PaymentMethod.
func (c *Service) Delete(ctx context.Context, pm payment_method.PaymentMethod) error {
	err := c.storer.Delete(ctx, pm)
	if err != nil {
		return fmt.Errorf("payment method delete: %w", err)
	}

	return nil
}

// QueryByID finds the payment method by the specified ID.
func (c *Service) QueryByID(ctx context.Context, paymentMethodID uuid.UUID) (payment_method.PaymentMethod, error) {
	pm, err := c.storer.QueryByID(ctx, paymentMethodID)
	if err != nil {
		return payment_method.PaymentMethod{}, fmt.Errorf("query: paymentMethodID[%s]: %w", paymentMethodID, err)
	}

	return pm, nil
}

// Query retrieves a list of existing payment methods.
func (c *Service) Query(
	ctx context.Context,
	filter payment_method.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]payment_method.PaymentMethod, error) {
	pms, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("payment method query: %w", err)
	}

	return pms, nil
}

// QueryByAccountID retrieves payment methods for a specific account.
func (c *Service) QueryByAccountID(ctx context.Context, accountID uuid.UUID) ([]payment_method.PaymentMethod, error) {
	pms, err := c.storer.QueryByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("query: accountID[%s]: %w", accountID, err)
	}

	return pms, nil
}
