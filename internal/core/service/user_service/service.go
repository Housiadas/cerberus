// Package user_service it is the service of the user domain
package user_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

type Service struct {
	log     logger.Logger
	storer  user.Storer
	uuidGen uuidgen.Generator
	clock   clock.Clock
	hasher  hasher.Hasher
}

// New constructs the service.
func New(
	log logger.Logger,
	storer user.Storer,
	uuidGen uuidgen.Generator,
	clock clock.Clock,
	hasher hasher.Hasher,
) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
		clock:   clock,
		hasher:  hasher,
	}
}

// NewWithTx constructs a new internal value that will use the
// specified transaction in any store-related calls.
func (c *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := c.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("user transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
		clock:   c.clock,
		hasher:  c.hasher,
	}

	return &bus, nil
}

// Count returns the total number of users.
func (c *Service) Count(ctx context.Context, filter user.QueryFilter) (int, error) {
	count, err := c.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("users count: %w", err)
	}

	return count, nil
}
