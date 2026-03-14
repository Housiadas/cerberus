// Package plan_cache contains plan-related functionality with caching.
package plan_cache

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/plan"
	"github.com/Housiadas/cerberus/pkg/cachemetrics"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/redis"
	"github.com/google/uuid"
	"github.com/viccon/sturdyc"
)

const (
	capacity           = 10000
	numShards          = 10
	evictionPercentage = 10
	cacheName          = "plan_cache"
)

var ttl = 5 * time.Minute

// Store manages the set of APIs for plan data and caching.
type Store struct {
	storer plan.Storer
	log    logger.Logger
	cache  *sturdyc.Client[plan.Plan]
}

// NewStore constructs the api for data and caching access.
func NewStore(
	ctx context.Context,
	log logger.Logger,
	storer plan.Storer,
	red redis.Client,
) *Store {
	recorder, err := cachemetrics.NewMeterRecorder(cacheName)
	if err != nil {
		log.Error(ctx, "error initializing plan cache metrics", err)
	}

	ds := redis.NewDistributedStorage(red, ttl)

	opts := make([]sturdyc.Option, 0, 3)
	opts = append(opts, sturdyc.WithDistributedStorage(ds))
	opts = append(opts, sturdyc.WithMetrics(recorder))
	opts = append(opts, sturdyc.WithDistributedMetrics(recorder))

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[plan.Plan](capacity, numShards, ttl, evictionPercentage, opts...),
	}
}

// NewWithTx creates a new Store that uses the specified transaction.
// Transaction operations bypass the cache.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (plan.Storer, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("plan cache new with tx: %w", err)
	}

	return storer, nil
}

// Create inserts a new plan into the database.
func (s *Store) Create(ctx context.Context, pl plan.Plan) error {
	err := s.storer.Create(ctx, pl)
	if err != nil {
		return fmt.Errorf("plan cache create: %w", err)
	}

	return nil
}

// Update modifies an existing plan in the database and invalidates the cache.
func (s *Store) Update(ctx context.Context, pl plan.Plan) error {
	err := s.storer.Update(ctx, pl)
	if err != nil {
		return fmt.Errorf("plan cache update: %w", err)
	}

	s.cache.Delete(pl.ID().String())

	return nil
}

// Delete removes a plan from the database and invalidates the cache.
func (s *Store) Delete(ctx context.Context, pl plan.Plan) error {
	err := s.storer.Delete(ctx, pl)
	if err != nil {
		return fmt.Errorf("plan cache delete: %w", err)
	}

	s.cache.Delete(pl.ID().String())

	return nil
}

// Query retrieves a list of existing plans from the database.
func (s *Store) Query(
	ctx context.Context,
	filter plan.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]plan.Plan, error) {
	plans, err := s.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("plan cache query: %w", err)
	}

	return plans, nil
}

// QueryByID gets the specified plan, checking L1 -> L2 -> DB.
func (s *Store) QueryByID(ctx context.Context, planID uuid.UUID) (plan.Plan, error) {
	pl, err := s.cache.GetOrFetch(
		ctx,
		planID.String(),
		func(ctx context.Context) (plan.Plan, error) {
			return s.storer.QueryByID(ctx, planID)
		},
	)
	if err != nil {
		return plan.Plan{}, fmt.Errorf("plan cache query by id: %w", err)
	}

	return pl, nil
}
