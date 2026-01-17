// Package user_cache contains user-related functionality with caching.
package user_cache

import (
	"context"
	"fmt"
	"net/mail"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/uuid"
	"github.com/viccon/sturdyc"
)

// Store manages the set of APIs for user data and caching.
type Store struct {
	storer user.Storer
	log    *logger.Service
	cache  *sturdyc.Client[user.User]
}

// NewStore constructs the api for data and caching access.
func NewStore(log *logger.Service, storer user.Storer, ttl time.Duration) *Store {
	const (
		capacity           = 10000
		numShards          = 10
		evictionPercentage = 10
	)

	return &Store{
		log:    log,
		storer: storer,
		cache:  sturdyc.New[user.User](capacity, numShards, ttl, evictionPercentage),
	}
}

// Query retrieves a list of existing users from the database.
func (s *Store) Query(
	ctx context.Context,
	filter user.QueryFilter,
	orderBy order.By,
	page web.Page,
) ([]user.User, error) {
	query, err := s.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("user query: %w", err)
	}

	return query, nil
}

// Count returns the total number of cards in the DB.
func (s *Store) Count(ctx context.Context, filter user.QueryFilter) (int, error) {
	count, err := s.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("user count: %w", err)
	}

	return count, nil
}

// QueryByID gets the specified user from the database.
func (s *Store) QueryByID(ctx context.Context, userID uuid.UUID) (user.User, error) {
	cachedUsr, ok := s.readCache(userID.String())
	if ok {
		return cachedUsr, nil
	}

	usr, err := s.storer.QueryByID(ctx, userID)
	if err != nil {
		return user.User{}, fmt.Errorf("user query by id: %w", err)
	}

	s.writeCache(usr)

	return usr, nil
}

// QueryByEmail gets the specified user from the database by email.
func (s *Store) QueryByEmail(ctx context.Context, email mail.Address) (user.User, error) {
	cachedUsr, ok := s.readCache(email.Address)
	if ok {
		return cachedUsr, nil
	}

	usr, err := s.storer.QueryByEmail(ctx, email)
	if err != nil {
		return user.User{}, fmt.Errorf("user query by email: %w", err)
	}

	s.writeCache(usr)

	return usr, nil
}

// readCache performs a safe search in the cache for the specified key.
func (s *Store) readCache(key string) (user.User, bool) {
	usr, exists := s.cache.Get(key)
	if !exists {
		return user.User{}, false
	}

	return usr, true
}

// writeCache performs a safe writing to the cache for the specified user.
func (s *Store) writeCache(bus user.User) {
	s.cache.Set(bus.ID.String(), bus)
	s.cache.Set(bus.Email.Address, bus)
}

// deleteCache performs a safe removal from the cache for the specified user.
func (s *Store) deleteCache(bus user.User) {
	s.cache.Delete(bus.ID.String())
	s.cache.Delete(bus.Email.Address)
}
