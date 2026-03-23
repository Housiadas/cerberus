package account_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/account"
)

// Delete removes the specified account.
func (s *Service) Delete(ctx context.Context, acc account.Account) error {
	err := s.storer.Delete(ctx, acc)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
