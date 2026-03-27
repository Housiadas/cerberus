// Package account is the service of the account domain
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
	tx      pgsql.Beginner
}

// NewService constructs the service.
func NewService(
	log logger,
	storer Storer,
	uuidGen generator,
	clock clock,
	tx pgsql.Beginner,
) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
		clock:   clock,
		tx:      tx,
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
		tx:      s.tx,
	}

	return &svc, nil
}

// Create adds a new Account to the system within a transaction.
func (s *Service) Create(ctx context.Context, na NewAccount) (Account, error) {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return Account{}, fmt.Errorf("uuid v7 error: %w", err)
	}

	now := s.clock.Now()
	acc := New(id, na.Name, sql.NullString{}, now, now, nil)

	txErr := pgsql.RunInTx(ctx, s.log, s.tx, func(tran pgsql.CommitRollbacker) error {
		storerTx, err := s.storer.NewWithTx(tran)
		if err != nil {
			return fmt.Errorf("account tx: %w", err)
		}

		return storerTx.Create(ctx, acc)
	})
	if txErr != nil {
		return Account{}, fmt.Errorf("create account: %w", txErr)
	}

	return acc, nil
}

// Update modifies information about an account within a transaction.
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

	txErr := pgsql.RunInTx(ctx, s.log, s.tx, func(tran pgsql.CommitRollbacker) error {
		storerTx, err := s.storer.NewWithTx(tran)
		if err != nil {
			return fmt.Errorf("account tx: %w", err)
		}

		return storerTx.Update(ctx, acc)
	})
	if txErr != nil {
		return Account{}, fmt.Errorf("update account: %w", txErr)
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

// Delete removes the specified account within a transaction.
func (s *Service) Delete(ctx context.Context, acc Account) error {
	txErr := pgsql.RunInTx(ctx, s.log, s.tx, func(tran pgsql.CommitRollbacker) error {
		storerTx, err := s.storer.NewWithTx(tran)
		if err != nil {
			return fmt.Errorf("account tx: %w", err)
		}

		return storerTx.Delete(ctx, acc)
	})
	if txErr != nil {
		return fmt.Errorf("delete account: %w", txErr)
	}

	return nil
}
