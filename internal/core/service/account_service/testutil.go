package account_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/account"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
)

// TestNewAccounts is a helper method for testing.
func TestNewAccounts(n int) []account.NewAccount {
	newAccounts := make([]account.NewAccount, n)

	for i := range n {
		na := account.NewAccount{
			Name: name.MustParse(fmt.Sprintf("Account%d", i)),
			Type: account.MustParse("personal"),
		}

		newAccounts[i] = na
	}

	return newAccounts
}

// TestSeedAccounts is a helper method for testing.
func TestSeedAccounts(ctx context.Context, n int, service *Service) ([]account.Account, error) {
	newAccounts := TestNewAccounts(n)

	accounts := make([]account.Account, len(newAccounts))

	for i, na := range newAccounts {
		acct, err := service.Create(ctx, na)
		if err != nil {
			return nil, fmt.Errorf("seeding account: idx: %d : %w", i, err)
		}

		accounts[i] = acct
	}

	return accounts, nil
}
