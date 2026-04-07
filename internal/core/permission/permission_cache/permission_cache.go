// Package permission_cache contains permission-related functionality with caching.
package permission_cache

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/sdk/distributed_storage"
	"github.com/Housiadas/cerberus/pkg/cachemetrics"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
	redisPck "github.com/redis/go-redis/v9"
	"github.com/viccon/sturdyc"
)

const (
	capacity           = 10000
	numShards          = 10
	evictionPercentage = 10
	cacheName          = "permission_cache"
)

var ttl = 5 * time.Minute

// Client defines the interface for Redis operations used by distributed storage.
type redisClient interface {
	Get(ctx context.Context, key string) *redisPck.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redisPck.StatusCmd
	MGet(ctx context.Context, keys ...string) *redisPck.SliceCmd
	Pipeline() redisPck.Pipeliner
}

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
	red redisClient,
) *Store {
	recorder, err := cachemetrics.NewMeterRecorder(cacheName)
	if err != nil {
		log.Error(ctx, "error initializing permission cache metrics", err)
	}

	ds := distributed_storage.New(red, ttl)

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
	cur cursor.Cursor,
) ([]permission.Permission, error) {
	perms, err := s.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("permission cache query: %w", err)
	}

	return perms, nil
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
