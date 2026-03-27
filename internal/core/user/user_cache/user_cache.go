// Package user_cache contains user-related functionality with caching.
package user_cache

import (
	"context"
	"fmt"
	"net/mail"
	"time"

	"github.com/Housiadas/cerberus/internal/core/user"
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
	cacheName          = "user_cache"
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

// Storer interface declares the behavior this package needs to persist and retrieve data.
type storer interface {
	NewWithTx(tx pgsql.CommitRollbacker) (user.Storer, error)
	Create(ctx context.Context, usr user.User) error
	Update(ctx context.Context, usr user.User) error
	Delete(ctx context.Context, usr user.User) error
	Query(
		ctx context.Context,
		filter user.QueryFilter,
		orderBy order.By,
		cur cursor.Cursor,
	) ([]user.User, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (user.User, error)
	QueryByEmail(ctx context.Context, email mail.Address) (user.User, error)
}

// Store manages the set of APIs for user data and caching.
type Store struct {
	storer storer
	log    logger
	cache  *sturdyc.Client[user.User]
}

// NewStore constructs the api for data and caching access.
func NewStore(
	ctx context.Context,
	log logger,
	storer storer,
	red redisClient,
) *Store {
	// Wire up OTel metrics. Errors are non-fatal: the cache works without metrics.
	recorder, err := cachemetrics.NewMeterRecorder(cacheName)
	if err != nil {
		log.Error(ctx, "error initializing user cache metrics", err)
	}

	ds := distributed_storage.New(red, ttl)

	opts := make([]sturdyc.Option, 0, 6)
	opts = append(opts, sturdyc.WithDistributedStorage(ds))
	opts = append(opts, sturdyc.WithMetrics(recorder))
	opts = append(opts, sturdyc.WithDistributedMetrics(recorder))

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[user.User](capacity, numShards, ttl, evictionPercentage, opts...),
	}
}

// NewWithTx creates a new Store that uses the specified transaction.
// Transaction operations bypass the cache.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (user.Storer, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("user cache new with tx: %w", err)
	}

	return storer, nil
}

// Create inserts a new user into the database.
func (s *Store) Create(ctx context.Context, usr user.User) error {
	err := s.storer.Create(ctx, usr)
	if err != nil {
		return fmt.Errorf("user cache create: %w", err)
	}

	return nil
}

// Update modifies an existing user in the database and invalidates the cache.
func (s *Store) Update(ctx context.Context, usr user.User) error {
	err := s.storer.Update(ctx, usr)
	if err != nil {
		return fmt.Errorf("user cache update: %w", err)
	}

	s.cache.Delete(usr.ID().String())
	s.cache.Delete(usr.Email().Address)

	return nil
}

// Delete removes a user from the database and invalidates the cache.
func (s *Store) Delete(ctx context.Context, usr user.User) error {
	err := s.storer.Delete(ctx, usr)
	if err != nil {
		return fmt.Errorf("user cache delete: %w", err)
	}

	s.cache.Delete(usr.ID().String())
	s.cache.Delete(usr.Email().Address)

	return nil
}

// Query retrieves a list of existing users from the database.
func (s *Store) Query(
	ctx context.Context,
	filter user.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]user.User, error) {
	users, err := s.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("user cache query: %w", err)
	}

	return users, nil
}

// QueryByID gets the specified user, checking L1 -> L2 -> DB.
func (s *Store) QueryByID(ctx context.Context, userID uuid.UUID) (user.User, error) {
	usr, err := s.cache.GetOrFetch(
		ctx,
		userID.String(),
		func(ctx context.Context) (user.User, error) {
			return s.storer.QueryByID(ctx, userID)
		},
	)
	if err != nil {
		return user.User{}, fmt.Errorf("user cache query by id: %w", err)
	}

	return usr, nil
}

// QueryByEmail gets the specified user by email, checking L1 -> L2 -> DB.
func (s *Store) QueryByEmail(ctx context.Context, email mail.Address) (user.User, error) {
	usr, err := s.cache.GetOrFetch(
		ctx,
		email.Address,
		func(ctx context.Context) (user.User, error) {
			return s.storer.QueryByEmail(ctx, email)
		},
	)
	if err != nil {
		return user.User{}, fmt.Errorf("user cache query by email: %w", err)
	}

	return usr, nil
}
