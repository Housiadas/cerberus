// Package role_cache contains role-related functionality with caching.
package role_cache

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
	"github.com/redis/go-redis/v9"
	"github.com/viccon/sturdyc"
)

const (
	capacity           = 10000
	numShards          = 10
	evictionPercentage = 10
	cacheName          = "role_cache"
)

var ttl = 5 * time.Minute

// Client defines the interface for Redis operations used by distributed storage.
type redisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
	Pipeline() redis.Pipeliner
}

type storer interface {
	CreateRole(ctx context.Context, arg db.CreateRoleParams) (db.Role, error)
	UpdateRole(ctx context.Context, arg db.UpdateRoleParams) (db.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	QueryRoles(
		ctx context.Context,
		filter db.RoleQueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]db.Role, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (db.Role, error)
}

// Store manages the set of APIs for role data and caching.
type Store struct {
	storer storer
	log    logger.Logger
	cache  *sturdyc.Client[db.Role]
}

// NewStore constructs the api for data and caching access.
func NewStore(
	log logger.Logger,
	storer storer,
	red redisClient,
) *Store {
	recorder, err := cachemetrics.NewMeterRecorder(cacheName)
	if err != nil {
		log.Error(context.Background(), "error initializing role cache metrics", err)
	}

	ds := distributed_storage.New(red, ttl)

	opts := make([]sturdyc.Option, 0, 3)
	opts = append(opts, sturdyc.WithDistributedStorage(ds))
	opts = append(opts, sturdyc.WithMetrics(recorder))
	opts = append(opts, sturdyc.WithDistributedMetrics(recorder))

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[db.Role](capacity, numShards, ttl, evictionPercentage, opts...),
	}
}

// CreateRole inserts a new role into the database.
func (s *Store) CreateRole(
	ctx context.Context,
	arg db.CreateRoleParams,
) (db.Role, error) {
	dbRole, err := s.storer.CreateRole(ctx, arg)
	if err != nil {
		return db.Role{}, fmt.Errorf("role cache create: %w", err)
	}

	return dbRole, nil
}

// UpdateRole modifies an existing role in the database and invalidates the cache.
func (s *Store) UpdateRole(
	ctx context.Context,
	arg db.UpdateRoleParams,
) (db.Role, error) {
	dbRole, err := s.storer.UpdateRole(ctx, arg)
	if err != nil {
		return db.Role{}, fmt.Errorf("role cache update: %w", err)
	}

	s.cache.Delete(dbRole.ID.String())

	return dbRole, nil
}

// DeleteRole removes a role from the database and invalidates the cache.
func (s *Store) DeleteRole(ctx context.Context, id uuid.UUID) error {
	err := s.storer.DeleteRole(ctx, id)
	if err != nil {
		return fmt.Errorf("role cache delete: %w", err)
	}

	s.cache.Delete(id.String())

	return nil
}

// QueryRoles retrieves a list of existing roles from the database.
func (s *Store) QueryRoles(
	ctx context.Context,
	filter db.RoleQueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]db.Role, error) {
	roles, err := s.storer.QueryRoles(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("role cache query: %w", err)
	}

	return roles, nil
}

// GetRoleByID gets the specified role, checking L1 -> L2 -> DB.
func (s *Store) GetRoleByID(ctx context.Context, id uuid.UUID) (db.Role, error) {
	rl, err := s.cache.GetOrFetch(
		ctx,
		id.String(),
		func(ctx context.Context) (db.Role, error) {
			return s.storer.GetRoleByID(ctx, id)
		},
	)
	if err != nil {
		return db.Role{}, fmt.Errorf("role cache query by id: %w", err)
	}

	return rl, nil
}
