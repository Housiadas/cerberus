package account_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/account"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

// Query retrieves a list of existing accounts.
func (s *Service) Query(
	ctx context.Context,
	filter account.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]account.Account, error) {
	accounts, err := s.storer.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return accounts, nil
}

// QueryByID finds the account by the specified ID.
func (s *Service) QueryByID(ctx context.Context, accountID uuid.UUID) (account.Account, error) {
	acc, err := s.storer.QueryByID(ctx, accountID)
	if err != nil {
		return account.Account{}, fmt.Errorf("query: accountID[%s]: %w", accountID, err)
	}

	return acc, nil
}
