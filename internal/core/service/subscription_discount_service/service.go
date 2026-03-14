// Package subscription_discount_service provides internal access to subscription discount core.
package subscription_discount_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/subscription_discount"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for subscription discount access.
type Service struct {
	log     logger.Logger
	storer  subscription_discount.Storer
	uuidGen uuidgen.Generator
}

// New constructor.
func New(
	log logger.Logger,
	storer subscription_discount.Storer,
	uuidGen uuidgen.Generator,
) *Service {
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
		return nil, fmt.Errorf("subscription discount transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new subscription_discount.SubscriptionDiscount to the system.
func (c *Service) Create(
	ctx context.Context,
	nsd subscription_discount.NewSubscriptionDiscount,
) (subscription_discount.SubscriptionDiscount, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return subscription_discount.SubscriptionDiscount{}, fmt.Errorf(
			"subscription discount uuid generate: %w",
			err,
		)
	}

	now := time.Now()
	sd := subscription_discount.New(id, nsd.SubscriptionID, nsd.CouponID, now, now)

	err = c.storer.Create(ctx, sd)
	if err != nil {
		return subscription_discount.SubscriptionDiscount{}, fmt.Errorf(
			"subscription discount create: %w",
			err,
		)
	}

	return sd, nil
}

// QueryByID finds the subscription discount by the specified ID.
func (c *Service) QueryByID(
	ctx context.Context,
	subscriptionDiscountID uuid.UUID,
) (subscription_discount.SubscriptionDiscount, error) {
	sd, err := c.storer.QueryByID(ctx, subscriptionDiscountID)
	if err != nil {
		return subscription_discount.SubscriptionDiscount{}, fmt.Errorf(
			"query: subscriptionDiscountID[%s]: %w",
			subscriptionDiscountID,
			err,
		)
	}

	return sd, nil
}

// Query retrieves a list of existing subscription discounts.
func (c *Service) Query(
	ctx context.Context,
	filter subscription_discount.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]subscription_discount.SubscriptionDiscount, error) {
	sds, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("subscription discount query: %w", err)
	}

	return sds, nil
}

// QueryBySubscriptionID retrieves subscription discounts for a specific subscription.
func (c *Service) QueryBySubscriptionID(
	ctx context.Context,
	subscriptionID uuid.UUID,
) ([]subscription_discount.SubscriptionDiscount, error) {
	sds, err := c.storer.QueryBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("query: subscriptionID[%s]: %w", subscriptionID, err)
	}

	return sds, nil
}
