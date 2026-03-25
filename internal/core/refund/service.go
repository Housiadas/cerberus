// Package refund_service is the service of the refund domain
package refund

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

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

// Service manages refund operations.
type Service struct {
	log    logger
	storer Storer
}

// NewService constructs the service.
func NewService(log logger, storer Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new Service that uses the specified transaction.
func (s *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := s.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("refund transaction issue: %w", err)
	}

	return &Service{log: s.log, storer: storer}, nil
}

// Create adds a new refund to the system.
func (s *Service) Create(ctx context.Context, ref Refund) error {
	err := s.storer.Create(ctx, ref)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

// QueryByAccountID retrieves refunds for an account.
func (s *Service) QueryByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
) ([]Refund, error) {
	refs, err := s.storer.QueryByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("query by account id: %w", err)
	}

	return refs, nil
}
