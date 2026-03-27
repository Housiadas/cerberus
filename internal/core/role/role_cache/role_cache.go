// Package role_cache contains role-related functionality with caching.
package role_cache

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/role"
	"github.com/Housiadas/cerberus/internal/distributed_storage"
	"github.com/Housiadas/cerberus/pkg/cachemetrics"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
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

type logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Debugc(ctx context.Context, caller int, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Infoc(ctx context.Context, caller int, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Warnc(ctx context.Context, caller int, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Errorc(ctx context.Context, caller int, msg string, args ...any)
}

// Client defines the interface for Redis operations used by distributed storage.
type redisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
	Pipeline() redis.Pipeliner
}

type storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (role.Storer, error)
	Create(ctx context.Context, role role.Role) error
	Update(ctx context.Context, role role.Role) error
	Delete(ctx context.Context, role role.Role) error
	Query(
		ctx context.Context,
		filter role.QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]role.Role, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (role.Role, error)
}

// Store manages the set of APIs for role data and caching.
type Store struct {
	storer storer
	log    logger
	cache  *sturdyc.Client[role.Role]
}

// NewStore constructs the api for data and caching access.
func NewStore(
	ctx context.Context,
	log logger,
	storer role.Storer,
	red redisClient,
) *Store {
	recorder, err := cachemetrics.NewMeterRecorder(cacheName)
	if err != nil {
		log.Error(ctx, "error initializing role cache metrics", err)
	}

	ds := distributed_storage.New(red, ttl)

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
