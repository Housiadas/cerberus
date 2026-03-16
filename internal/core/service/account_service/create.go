package account_service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/account"
)

// Create adds a new Account to the system.
func (s *Service) Create(ctx context.Context, na account.NewAccount) (account.Account, error) {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return account.Account{}, fmt.Errorf("uuid v7 error: %w", err)
	}

	now := s.clock.Now()
	acc := account.New(id, na.Name, sql.NullString{}, now, now, nil)

	err = s.storer.Create(ctx, acc)
	if err != nil {
		return account.Account{}, fmt.Errorf("create: %w", err)
	}

	return acc, nil
}
