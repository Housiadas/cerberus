// Package account_service is the service of the account domain
package account

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
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

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}

type Service struct {
	log     logger
	storer  Storer
	uuidGen generator
	clock   clock
}

// NewService constructs the service.
func NewService(
	log logger,
	storer Storer,
	uuidGen generator,
	clock clock,
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

// Create adds a new Account to the system.
func (s *Service) Create(ctx context.Context, na NewAccount) (Account, error) {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return Account{}, fmt.Errorf("uuid v7 error: %w", err)
	}

	now := s.clock.Now()
	acc := New(id, na.Name, sql.NullString{}, now, now, nil)

	err = s.storer.Create(ctx, acc)
	if err != nil {
		return Account{}, fmt.Errorf("create: %w", err)
	}

	return acc, nil
}

// Update modifies information about an account.
func (s *Service) Update(
	ctx context.Context,
	acc Account,
	ua UpdateAccount,
) (Account, error) {
	if ua.Name != nil {
		acc = acc.WithName(*ua.Name)
	}

	if ua.StripeCustomerID != nil {
		acc = acc.WithStripeCustomerID(*ua.StripeCustomerID)
	}

	acc = acc.WithUpdatedAt(s.clock.Now())

	err := s.storer.Update(ctx, acc)
	if err != nil {
		return Account{}, fmt.Errorf("update: %w", err)
	}

	return acc, nil
}

// Query retrieves a list of existing accounts.
func (s *Service) Query(
	ctx context.Context,
	filter QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]Account, error) {
	accounts, err := s.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return accounts, nil
}

// QueryByID finds the account by the specified ID.
func (s *Service) QueryByID(ctx context.Context, accountID uuid.UUID) (Account, error) {
	acc, err := s.storer.QueryByID(ctx, accountID)
	if err != nil {
		return Account{}, fmt.Errorf("query: accountID[%s]: %w", accountID, err)
	}

	return acc, nil
}

// Delete removes the specified account.
func (s *Service) Delete(ctx context.Context, acc Account) error {
	err := s.storer.Delete(ctx, acc)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
