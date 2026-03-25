// Package account_service is the service of the account domain
package account_service

import (
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

type Service struct {
	log     logger.Logger
	storer  account.Storer
	uuidGen uuidgen.Generator
	clock   clock.Clock
}

// New constructs the service.
func New(
	log logger.Logger,
	storer account.Storer,
	uuidGen uuidgen.Generator,
	clock clock.Clock,
) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
		clock:   clock,
	}
}

// NewWithTx constructs a new internal value that will use the
// specified transaction in any store-related calls.
func (s *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("account transaction issue: %w", err)
	}

	svc := Service{
		log:     s.log,
		storer:  storer,
		uuidGen: s.uuidGen,
		clock:   s.clock,
	}

	return &svc, nil
}
