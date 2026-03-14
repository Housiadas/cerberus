// Package subscription_service provides internal access to subscription core.
package subscription_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/subscription"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for subscription access.
type Service struct {
	log     logger.Logger
	storer  subscription.Storer
	uuidGen uuidgen.Generator
}

// New constructor.
func New(log logger.Logger, storer subscription.Storer, uuidGen uuidgen.Generator) *Service {
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
		return nil, fmt.Errorf("subscription transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new subscription.Subscription to the system.
func (c *Service) Create(
	ctx context.Context,
	ns subscription.NewSubscription,
) (subscription.Subscription, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("subscription uuid generate: %w", err)
	}

	now := time.Now()
	sub := subscription.New(
		id,
		ns.AccountID,
		ns.PlanID,
		ns.Status,
		ns.CurrentPeriodStart,
		ns.CurrentPeriodEnd,
		now,
		now,
	)

	err = c.storer.Create(ctx, sub)
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("subscription create: %w", err)
	}

	return sub, nil
}

// Update modifies information about a subscription.Subscription.
func (c *Service) Update(
	ctx context.Context,
	sub subscription.Subscription,
	us subscription.UpdateSubscription,
) (subscription.Subscription, error) {
	if us.Status != nil {
		sub = sub.WithStatus(*us.Status)
	}

	if us.CurrentPeriodStart != nil {
		sub = sub.WithCurrentPeriodStart(*us.CurrentPeriodStart)
	}

	if us.CurrentPeriodEnd != nil {
		sub = sub.WithCurrentPeriodEnd(*us.CurrentPeriodEnd)
	}

	sub = sub.WithUpdatedAt(time.Now())

	err := c.storer.Update(ctx, sub)
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("subscription update: %w", err)
	}

	return sub, nil
}

// QueryByID finds the subscription by the specified ID.
func (c *Service) QueryByID(
	ctx context.Context,
	subscriptionID uuid.UUID,
) (subscription.Subscription, error) {
	sub, err := c.storer.QueryByID(ctx, subscriptionID)
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf(
			"query: subscriptionID[%s]: %w",
			subscriptionID,
			err,
		)
	}

	return sub, nil
}

// QueryByAccountID finds subscriptions by account ID.
func (c *Service) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]subscription.Subscription, error) {
	subs, err := c.storer.QueryByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("query: accountID[%s]: %w", accountID, err)
	}

	return subs, nil
}

// Query retrieves a list of existing subscriptions.
func (c *Service) Query(
	ctx context.Context,
	filter subscription.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]subscription.Subscription, error) {
	subs, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("subscription query: %w", err)
	}

	return subs, nil
}
