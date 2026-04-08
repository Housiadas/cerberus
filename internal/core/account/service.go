// Package account is the service of the account domain
package account

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

type Service struct {
	log     logger.Logger
	storer  Storer
	uuidGen generator
	clock   clock
	tx      transactor
}

// NewService constructs the service.
func NewService(
	log logger.Logger,
	storer Storer,
	uuidGen generator,
	clock clock,
	tx transactor,
) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
		clock:   clock,
		tx:      tx,
	}
}

// Create adds a new Account to the system within a transaction.
func (s *Service) Create(ctx context.Context, na NewAccount) (Account, error) {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return Account{}, fmt.Errorf("uuid v7 error: %w", err)
	}

	now := s.clock.Now()
	acc := New(id, na.Name, sql.NullString{}, now, now, nil)

	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		return s.storer.Create(txCtx, acc)
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

	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		return s.storer.Update(txCtx, acc)
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
	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		return s.storer.Delete(txCtx, acc)
	})
	if txErr != nil {
		return fmt.Errorf("delete account: %w", txErr)
	}

	return nil
}
