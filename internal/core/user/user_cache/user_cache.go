// Package user_cache contains user-related functionality with caching.
package user_cache

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
	cacheName          = "user_cache"
)

var ttl = 5 * time.Minute

// Client defines the interface for Redis operations used by distributed storage.
type redisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
	Pipeline() redis.Pipeliner
}

// storer interface declares the behavior this package needs to persist and retrieve data.
type storer interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	QueryUsers(
		ctx context.Context,
		filter db.UserQueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]db.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
}

// Store manages the set of APIs for user data and caching.
type Store struct {
	storer storer
	log    logger.Logger
	cache  *sturdyc.Client[db.User]
}

// NewStore constructs the api for data and caching access.
func NewStore(
	log logger.Logger,
	storer storer,
	red redisClient,
) *Store {
	// Wire up OTel metrics. Errors are non-fatal: the cache works without metrics.
	recorder, err := cachemetrics.NewMeterRecorder(cacheName)
	if err != nil {
		log.Error(context.Background(), "error initializing user cache metrics", err)
	}

	ds := distributed_storage.New(red, ttl)

	opts := make([]sturdyc.Option, 0, 6)
	opts = append(opts, sturdyc.WithDistributedStorage(ds))
	opts = append(opts, sturdyc.WithMetrics(recorder))
	opts = append(opts, sturdyc.WithDistributedMetrics(recorder))

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[db.User](capacity, numShards, ttl, evictionPercentage, opts...),
	}
}

// CreateUser inserts a new user into the database.
func (s *Store) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	dbUser, err := s.storer.CreateUser(ctx, arg)
	if err != nil {
		return db.User{}, fmt.Errorf("user cache create: %w", err)
	}

	return dbUser, nil
}

// UpdateUser modifies an existing user in the database and invalidates the cache.
func (s *Store) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	dbUser, err := s.storer.UpdateUser(ctx, arg)
	if err != nil {
		return db.User{}, fmt.Errorf("user cache update: %w", err)
	}

	s.cache.Delete(dbUser.ID.String())
	s.cache.Delete(dbUser.Email)

	return dbUser, nil
}

// DeleteUser removes a user from the database and invalidates the cache.
func (s *Store) DeleteUser(ctx context.Context, id uuid.UUID) error {
	dbUser, err := s.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found cache delete: %w", err)
	}

	err = s.storer.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("user cache delete: %w", err)
	}

	s.cache.Delete(dbUser.ID.String())
	s.cache.Delete(dbUser.Email)

	return nil
}

// QueryUsers retrieves a list of existing users from the database.
func (s *Store) QueryUsers(
	ctx context.Context,
	filter db.UserQueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]db.User, error) {
	users, err := s.storer.QueryUsers(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("user cache query: %w", err)
	}

	return users, nil
}

// GetUserByID gets the specified user, checking L1 -> L2 -> DB.
func (s *Store) GetUserByID(
	ctx context.Context,
	id uuid.UUID,
) (db.User, error) {
	usr, err := s.cache.GetOrFetch(
		ctx,
		id.String(),
		func(ctx context.Context) (db.User, error) {
			return s.storer.GetUserByID(ctx, id)
		},
	)
	if err != nil {
		return db.User{}, fmt.Errorf("user cache query by id: %w", err)
	}

	return usr, nil
}

// GetUserByEmail gets the specified user by email, checking L1 -> L2 -> DB.
func (s *Store) GetUserByEmail(
	ctx context.Context,
	email string,
) (db.User, error) {
	usr, err := s.cache.GetOrFetch(
		ctx,
		email,
		func(ctx context.Context) (db.User, error) {
			return s.storer.GetUserByEmail(ctx, email)
		},
	)
	if err != nil {
		return db.User{}, fmt.Errorf("user cache query by email: %w", err)
	}

	return usr, nil
}
