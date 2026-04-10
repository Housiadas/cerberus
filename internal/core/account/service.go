// Package account is the service of the account domain
package account

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

type Service struct {
	log     logger.Logger
	storer  storer
	uuidGen generator
	clock   clock
	tx      transactor
}

// NewService constructs the service.
func NewService(
	log logger.Logger,
	storer storer,
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
	params := toCreateAccountParams(id, na, now)

	var created Account
	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		dbAcc, err := s.storer.CreateAccount(txCtx, params)
		if err != nil {
			return err
		}

		created = toDomainAccount(dbAcc)

		return nil
	})
	if txErr != nil {
		return Account{}, fmt.Errorf("create account: %w", txErr)
	}

	return created, nil
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

	params := toUpdateAccountParams(acc)

	var updated Account
	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		dbAcc, err := s.storer.UpdateAccount(txCtx, params)
		if err != nil {
			return err
		}

		updated = toDomainAccount(dbAcc)

		return nil
	})
	if txErr != nil {
		return Account{}, fmt.Errorf("update account: %w", txErr)
	}

	return updated, nil
}

// Query retrieves a list of existing accounts.
func (s *Service) Query(
	ctx context.Context,
	filter QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]Account, error) {
	dbFilter := toDBQueryFilter(filter)

	dbAccounts, err := s.storer.QueryAccounts(ctx, dbFilter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return toDomainAccounts(dbAccounts), nil
}

// QueryByID finds the account by the specified ID.
func (s *Service) QueryByID(ctx context.Context, accountID uuid.UUID) (Account, error) {
	dbAcc, err := s.storer.GetAccountByID(ctx, accountID)
	if err != nil {
		return Account{}, fmt.Errorf("query: accountID[%s]: %w", accountID, err)
	}

	return toDomainAccountFromGetByID(dbAcc), nil
}

// Delete removes the specified account within a transaction.
func (s *Service) Delete(ctx context.Context, acc Account) error {
	txErr := s.tx.RunInTx(ctx, func(txCtx context.Context) error {
		return s.storer.DeleteAccount(txCtx, acc.ID())
	})
	if txErr != nil {
		return fmt.Errorf("delete account: %w", txErr)
	}

	return nil
}
