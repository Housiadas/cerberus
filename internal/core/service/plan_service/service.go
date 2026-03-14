// Package plan_service provides internal access to plan core.
package plan_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/plan"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for plan access.
type Service struct {
	log     logger.Logger
	storer  plan.Storer
	uuidGen uuidgen.Generator
}

// New constructor.
func New(log logger.Logger, storer plan.Storer, uuidGen uuidgen.Generator) *Service {
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
		return nil, fmt.Errorf("plan transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new plan.Plan to the system.
func (c *Service) Create(ctx context.Context, np plan.NewPlan) (plan.Plan, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return plan.Plan{}, fmt.Errorf("plan uuid generate: %w", err)
	}

	now := time.Now()
	pl := plan.New(id, np.Name, np.Description, np.Interval, np.PriceCents, np.Currency, true, now, now)

	err = c.storer.Create(ctx, pl)
	if err != nil {
		return plan.Plan{}, fmt.Errorf("plan create: %w", err)
	}

	return pl, nil
}

// Update modifies information about a plan.Plan.
func (c *Service) Update(
	ctx context.Context,
	pl plan.Plan,
	up plan.UpdatePlan,
) (plan.Plan, error) {
	if up.Name != nil {
		pl = pl.WithName(*up.Name)
	}
	if up.Description != nil {
		pl = pl.WithDescription(*up.Description)
	}
	if up.Interval != nil {
		pl = pl.WithInterval(*up.Interval)
	}
	if up.PriceCents != nil {
		pl = pl.WithPriceCents(*up.PriceCents)
	}
	if up.Currency != nil {
		pl = pl.WithCurrency(*up.Currency)
	}
	if up.IsActive != nil {
		pl = pl.WithIsActive(*up.IsActive)
	}

	pl = pl.WithUpdatedAt(time.Now())

	err := c.storer.Update(ctx, pl)
	if err != nil {
		return plan.Plan{}, fmt.Errorf("plan update: %w", err)
	}

	return pl, nil
}

// Delete removes the specified plan.Plan.
func (c *Service) Delete(ctx context.Context, pl plan.Plan) error {
	err := c.storer.Delete(ctx, pl)
	if err != nil {
		return fmt.Errorf("plan delete: %w", err)
	}

	return nil
}

// QueryByID finds the plan by the specified ID.
func (c *Service) QueryByID(ctx context.Context, planID uuid.UUID) (plan.Plan, error) {
	pl, err := c.storer.QueryByID(ctx, planID)
	if err != nil {
		return plan.Plan{}, fmt.Errorf("query: planID[%s]: %w", planID, err)
	}

	return pl, nil
}

// Query retrieves a list of existing plans.
func (c *Service) Query(
	ctx context.Context,
	filter plan.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]plan.Plan, error) {
	plans, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("plan query: %w", err)
	}

	return plans, nil
}
