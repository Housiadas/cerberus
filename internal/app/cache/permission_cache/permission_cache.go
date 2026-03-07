// Package permission_cache contains permission-related functionality with caching.
package permission_cache

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/Housiadas/cerberus/pkg/cachemetrics"
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
	cacheName          = "permission_cache"
)

var ttl = 5 * time.Minute

// Store manages the set of APIs for permission data and caching.
type Store struct {
	storer permission.Storer
	log    logger.Logger
	cache  *sturdyc.Client[permission.Permission]
}

// NewStore constructs the api for data and caching access.
func NewStore(
	ctx context.Context,
	log logger.Logger,
	storer permission.Storer,
	red redis.Client,
) *Store {
	recorder, err := cachemetrics.NewMeterRecorder(cacheName)
	if err != nil {
		log.Error(ctx, "error initializing permission cache metrics", err)
	}

	ds := redis.NewDistributedStorage(red, ttl)

	opts := make([]sturdyc.Option, 0, 3)
	opts = append(opts, sturdyc.WithDistributedStorage(ds))
	opts = append(opts, sturdyc.WithMetrics(recorder))
	opts = append(opts, sturdyc.WithDistributedMetrics(recorder))

	return &Store{
		log:    log,
		storer: storer,
		cache: sturdyc.New[permission.Permission](
			capacity,
			numShards,
			ttl,
			evictionPercentage,
			opts...),
	}
}

// NewWithTx creates a new Store that uses the specified transaction.
// Transaction operations bypass the cache.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (permission.Storer, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("permission cache new with tx: %w", err)
	}

	return storer, nil
}

// Create inserts a new permission into the database.
func (s *Store) Create(ctx context.Context, p permission.Permission) error {
	err := s.storer.Create(ctx, p)
	if err != nil {
		return fmt.Errorf("permission cache create: %w", err)
	}

	return nil
}

// Update modifies an existing permission in the database and invalidates the cache.
func (s *Store) Update(ctx context.Context, p permission.Permission) error {
	err := s.storer.Update(ctx, p)
	if err != nil {
		return fmt.Errorf("permission cache update: %w", err)
	}

	s.cache.Delete(p.ID().String())

	return nil
}

// Delete removes a permission from the database and invalidates the cache.
func (s *Store) Delete(ctx context.Context, p permission.Permission) error {
	err := s.storer.Delete(ctx, p)
	if err != nil {
		return fmt.Errorf("permission cache delete: %w", err)
	}

	s.cache.Delete(p.ID().String())

	return nil
}

// Query retrieves a list of existing permissions from the database.
func (s *Store) Query(
	ctx context.Context,
	filter permission.QueryFilter,
	orderBy order.By,
	page page.Page,
) ([]permission.Permission, error) {
	perms, err := s.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("permission cache query: %w", err)
	}

	return perms, nil
}

// Count returns the total number of permissions in the DB.
func (s *Store) Count(ctx context.Context, filter permission.QueryFilter) (int, error) {
	count, err := s.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("permission cache count: %w", err)
	}

	return count, nil
}

// QueryByID gets the specified permission, checking L1 -> L2 -> DB.
func (s *Store) QueryByID(
	ctx context.Context,
	permissionID uuid.UUID,
) (permission.Permission, error) {
	p, err := s.cache.GetOrFetch(
		ctx,
		permissionID.String(),
		func(ctx context.Context) (permission.Permission, error) {
			return s.storer.QueryByID(ctx, permissionID)
		},
	)
	if err != nil {
		return permission.Permission{}, fmt.Errorf("permission cache query by id: %w", err)
	}

	return p, nil
}
