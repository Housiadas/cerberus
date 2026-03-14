// Package account_service provides internal access to account core.
package account_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/account"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

// Service manages the set of APIs for account access.
type Service struct {
	log     logger.Logger
	storer  account.Storer
	uuidGen uuidgen.Generator
}

// New constructor.
func New(log logger.Logger, storer account.Storer, uuidGen uuidgen.Generator) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
	}
}

// NewWithTx constructs a new internal value that will use the
// specified transaction in any store-related calls.
func (c *Service) NewWithTx(tx pgsql.CommitRollbacker) (*Service, error) {
	storer, err := c.storer.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("account transaction issue: %w", err)
	}

	bus := Service{
		log:     c.log,
		storer:  storer,
		uuidGen: c.uuidGen,
	}

	return &bus, nil
}

// Create adds a new account.Account to the system.
func (c *Service) Create(ctx context.Context, na account.NewAccount) (account.Account, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return account.Account{}, fmt.Errorf("account uuid generate: %w", err)
	}

	now := time.Now()
	acct := account.New(id, na.Name, na.Type, true, now, now, nil)

	err = c.storer.Create(ctx, acct)
	if err != nil {
		return account.Account{}, fmt.Errorf("account create: %w", err)
	}

	return acct, nil
}

// Update modifies information about an account.Account.
func (c *Service) Update(
	ctx context.Context,
	acct account.Account,
	ua account.UpdateAccount,
) (account.Account, error) {
	if ua.Name != nil {
		acct = acct.WithName(*ua.Name)
	}
	if ua.Type != nil {
		acct = acct.WithType(*ua.Type)
	}
	if ua.Enabled != nil {
		acct = acct.WithEnabled(*ua.Enabled)
	}

	acct = acct.WithUpdatedAt(time.Now())

	err := c.storer.Update(ctx, acct)
	if err != nil {
		return account.Account{}, fmt.Errorf("account update: %w", err)
	}

	return acct, nil
}

// Delete removes the specified account.Account.
func (c *Service) Delete(ctx context.Context, acct account.Account) error {
	err := c.storer.Delete(ctx, acct)
	if err != nil {
		return fmt.Errorf("account delete: %w", err)
	}

	return nil
}

// QueryByID finds the account by the specified ID.
func (c *Service) QueryByID(ctx context.Context, accountID uuid.UUID) (account.Account, error) {
	acct, err := c.storer.QueryByID(ctx, accountID)
	if err != nil {
		return account.Account{}, fmt.Errorf("query: accountID[%s]: %w", accountID, err)
	}

	return acct, nil
}

// Query retrieves a list of existing accounts.
func (c *Service) Query(
	ctx context.Context,
	filter account.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]account.Account, error) {
	accounts, err := c.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("account query: %w", err)
	}

	return accounts, nil
}
