// Package role_cache contains role-related functionality with caching.
package role_cache

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/role"
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
	cacheName          = "role_cache"
)

var ttl = 5 * time.Minute

// Store manages the set of APIs for role data and caching.
type Store struct {
	storer role.Storer
	log    logger.Logger
	cache  *sturdyc.Client[role.Role]
}

// NewStore constructs the api for data and caching access.
func NewStore(
	ctx context.Context,
	log logger.Logger,
	storer role.Storer,
	red redis.Client,
) *Store {
	recorder, err := cachemetrics.NewMeterRecorder(cacheName)
	if err != nil {
		log.Error(ctx, "error initializing role cache metrics", err)
	}

	ds := redis.NewDistributedStorage(red, ttl)

	opts := make([]sturdyc.Option, 0, 3)
	opts = append(opts, sturdyc.WithDistributedStorage(ds))
	opts = append(opts, sturdyc.WithMetrics(recorder))
	opts = append(opts, sturdyc.WithDistributedMetrics(recorder))

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[role.Role](capacity, numShards, ttl, evictionPercentage, opts...),
	}
}

// NewWithTx creates a new Store that uses the specified transaction.
// Transaction operations bypass the cache.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (role.Storer, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("role cache new with tx: %w", err)
	}

	return storer, nil
}

// Create inserts a new role into the database.
func (s *Store) Create(ctx context.Context, rl role.Role) error {
	err := s.storer.Create(ctx, rl)
	if err != nil {
		return fmt.Errorf("role cache create: %w", err)
	}

	return nil
}

// Update modifies an existing role in the database and invalidates the cache.
func (s *Store) Update(ctx context.Context, rl role.Role) error {
	err := s.storer.Update(ctx, rl)
	if err != nil {
		return fmt.Errorf("role cache update: %w", err)
	}

	s.cache.Delete(rl.ID().String())

	return nil
}

// Delete removes a role from the database and invalidates the cache.
func (s *Store) Delete(ctx context.Context, rl role.Role) error {
	err := s.storer.Delete(ctx, rl)
	if err != nil {
		return fmt.Errorf("role cache delete: %w", err)
	}

	s.cache.Delete(rl.ID().String())

	return nil
}

// Query retrieves a list of existing roles from the database.
func (s *Store) Query(
	ctx context.Context,
	filter role.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]role.Role, error) {
	roles, err := s.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("role cache query: %w", err)
	}

	return roles, nil
}

// QueryByID gets the specified role, checking L1 -> L2 -> DB.
func (s *Store) QueryByID(ctx context.Context, roleID uuid.UUID) (role.Role, error) {
	rl, err := s.cache.GetOrFetch(
		ctx,
		roleID.String(),
		func(ctx context.Context) (role.Role, error) {
			return s.storer.QueryByID(ctx, roleID)
		},
	)
	if err != nil {
		return role.Role{}, fmt.Errorf("role cache query by id: %w", err)
	}

	return rl, nil
}
