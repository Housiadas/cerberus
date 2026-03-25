package account_usecase

import (
	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/pkg/clock"
)

// Account represents information about an account.
type Account struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	StripeCustomerID string `json:"stripeCustomerId,omitempty"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

func toAppAccount(acc account.Account) Account {
	return Account{
		ID:               acc.ID().String(),
		Name:             acc.Name(),
		StripeCustomerID: acc.StripeCustomerID().String,
		CreatedAt:        clock.Format(new(acc.CreatedAt())),
		UpdatedAt:        clock.Format(new(acc.UpdatedAt())),
	}
}

func toAppAccounts(accs []account.Account) []Account {
	app := make([]Account, len(accs))
	for i, acc := range accs {
		app[i] = toAppAccount(acc)
	}

	return app
}

// NewAccount defines the data needed to create a new account.
type NewAccount struct {
	Name string `json:"name" validate:"required"`
}

// UpdateAccount defines the data needed to update an account.
type UpdateAccount struct {
	Name *string `json:"name"`
}
