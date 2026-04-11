// Package permission_cache contains permission-related functionality with caching.
package permission_cache

import (
	"context"
	"fmt"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
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

type storer interface {
	CreatePermission(ctx context.Context, arg db.CreatePermissionParams) (db.Permission, error)
	UpdatePermission(ctx context.Context, arg db.UpdatePermissionParams) (db.Permission, error)
	DeletePermission(ctx context.Context, id uuid.UUID) error
	QueryPermissions(
		ctx context.Context,
		filter db.PermissionQueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]db.Permission, error)
	GetPermissionByID(ctx context.Context, id uuid.UUID) (db.Permission, error)
}

// Store manages the set of APIs for permission data and caching.
type Store struct {
	storer storer
	log    logger.Logger
	cache  *sturdyc.Client[db.Permission]
}

// NewStore constructs the api for data and caching access.
func NewStore(
	log logger.Logger,
	storer storer,
	red redisClient,
) *Store {
	recorder, err := cachemetrics.NewMeterRecorder(cacheName)
	if err != nil {
		log.Error(context.Background(), "error initializing permission cache metrics", err)
	}

	ds := distributed_storage.New(red, ttl)

	opts := make([]sturdyc.Option, 0, 3)
	opts = append(opts, sturdyc.WithDistributedStorage(ds))
	opts = append(opts, sturdyc.WithMetrics(recorder))
	opts = append(opts, sturdyc.WithDistributedMetrics(recorder))

	return &Store{
		log:    log,
		storer: storer,
		cache: sturdyc.New[db.Permission](
			capacity,
			numShards,
			ttl,
			evictionPercentage,
			opts...),
	}
}

// CreatePermission inserts a new permission into the database.
func (s *Store) CreatePermission(
	ctx context.Context,
	arg db.CreatePermissionParams,
) (db.Permission, error) {
	dbPermission, err := s.storer.CreatePermission(ctx, arg)
	if err != nil {
		return db.Permission{}, fmt.Errorf("permission cache create: %w", err)
	}

	return dbPermission, nil
}

// UpdatePermission modifies an existing permission in the database and invalidates the cache.
func (s *Store) UpdatePermission(
	ctx context.Context,
	arg db.UpdatePermissionParams,
) (db.Permission, error) {
	dbPermission, err := s.storer.UpdatePermission(ctx, arg)
	if err != nil {
		return db.Permission{}, fmt.Errorf("permission cache update: %w", err)
	}

	s.cache.Delete(dbPermission.ID.String())

	return dbPermission, nil
}

// DeletePermission removes a permission from the database and invalidates the cache.
func (s *Store) DeletePermission(ctx context.Context, id uuid.UUID) error {
	err := s.storer.DeletePermission(ctx, id)
	if err != nil {
		return fmt.Errorf("permission cache delete: %w", err)
	}

	s.cache.Delete(id.String())

	return nil
}

// QueryPermissions retrieves a list of existing permissions from the database.
func (s *Store) QueryPermissions(
	ctx context.Context,
	filter db.PermissionQueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]db.Permission, error) {
	perms, err := s.storer.QueryPermissions(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("permission cache query: %w", err)
	}

	return perms, nil
}

// GetPermissionByID gets the specified permission, checking L1 -> L2 -> DB.
func (s *Store) GetPermissionByID(
	ctx context.Context,
	id uuid.UUID,
) (db.Permission, error) {
	p, err := s.cache.GetOrFetch(
		ctx,
		id.String(),
		func(ctx context.Context) (db.Permission, error) {
			return s.storer.GetPermissionByID(ctx, id)
		},
	)
	if err != nil {
		return db.Permission{}, fmt.Errorf("permission cache query by id: %w", err)
	}

	return p, nil
}
