// Package account_cache contains account-related functionality with caching.
package account_cache

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/account"
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
	cacheName          = "account_cache"
)

var ttl = 5 * time.Minute

// Store manages the set of APIs for account data and caching.
type Store struct {
	storer account.Storer
	log    logger.Logger
	cache  *sturdyc.Client[account.Account]
}

// NewStore constructs the api for data and caching access.
func NewStore(
	ctx context.Context,
	log logger.Logger,
	storer account.Storer,
	red redis.Client,
) *Store {
	recorder, err := cachemetrics.NewMeterRecorder(cacheName)
	if err != nil {
		log.Error(ctx, "error initializing account cache metrics", err)
	}

	ds := redis.NewDistributedStorage(red, ttl)

	opts := make([]sturdyc.Option, 0, 3)
	opts = append(opts, sturdyc.WithDistributedStorage(ds))
	opts = append(opts, sturdyc.WithMetrics(recorder))
	opts = append(opts, sturdyc.WithDistributedMetrics(recorder))

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[account.Account](capacity, numShards, ttl, evictionPercentage, opts...),
	}
}

// NewWithTx creates a new Store that uses the specified transaction.
// Transaction operations bypass the cache.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (account.Storer, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("account cache new with tx: %w", err)
	}

	return storer, nil
}

// Create inserts a new account into the database.
func (s *Store) Create(ctx context.Context, acct account.Account) error {
	err := s.storer.Create(ctx, acct)
	if err != nil {
		return fmt.Errorf("account cache create: %w", err)
	}

	return nil
}

// Update modifies an existing account in the database and invalidates the cache.
func (s *Store) Update(ctx context.Context, acct account.Account) error {
	err := s.storer.Update(ctx, acct)
	if err != nil {
		return fmt.Errorf("account cache update: %w", err)
	}

	s.cache.Delete(acct.ID().String())

	return nil
}

// Delete removes an account from the database and invalidates the cache.
func (s *Store) Delete(ctx context.Context, acct account.Account) error {
	err := s.storer.Delete(ctx, acct)
	if err != nil {
		return fmt.Errorf("account cache delete: %w", err)
	}

	s.cache.Delete(acct.ID().String())

	return nil
}

// Query retrieves a list of existing accounts from the database.
func (s *Store) Query(
	ctx context.Context,
	filter account.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]account.Account, error) {
	accounts, err := s.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("account cache query: %w", err)
	}

	return accounts, nil
}

// QueryByID gets the specified account, checking L1 -> L2 -> DB.
func (s *Store) QueryByID(ctx context.Context, accountID uuid.UUID) (account.Account, error) {
	acct, err := s.cache.GetOrFetch(
		ctx,
		accountID.String(),
		func(ctx context.Context) (account.Account, error) {
			return s.storer.QueryByID(ctx, accountID)
		},
	)
	if err != nil {
		return account.Account{}, fmt.Errorf("account cache query by id: %w", err)
	}

	return acct, nil
}
