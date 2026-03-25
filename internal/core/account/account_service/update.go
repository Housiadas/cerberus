package account_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/account"
)

// Update modifies information about an account.
func (s *Service) Update(
	ctx context.Context,
	acc account.Account,
	ua account.UpdateAccount,
) (account.Account, error) {
	if ua.Name != nil {
		acc = acc.WithName(*ua.Name)
	}

	if ua.StripeCustomerID != nil {
		acc = acc.WithStripeCustomerID(*ua.StripeCustomerID)
	}

	acc = acc.WithUpdatedAt(s.clock.Now())

	err := s.storer.Update(ctx, acc)
	if err != nil {
		return account.Account{}, fmt.Errorf("update: %w", err)
	}

	return acc, nil
}
