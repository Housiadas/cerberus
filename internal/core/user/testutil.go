package user

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/internal/types/password"
	"github.com/google/uuid"
)

// TestSeedUsers is a helper method for testing.
func TestSeedUsers(ctx context.Context, n int, service *Service) ([]User, error) {
	newUsrs := testNewUsers(n)

	usrs := make([]User, len(newUsrs))

	for i, nu := range newUsrs {
		usr, err := service.Create(ctx, nu)
		if err != nil {
			return nil, fmt.Errorf("seeding user: idx: %d : %w", i, err)
		}

		usrs[i] = usr
	}

	return usrs, nil
}

// testNewUsers is a helper method for testing.
func testNewUsers(n int) []NewUser {
	newUsrs := make([]NewUser, n)

	suffix := uuid.New().String()[:8]

	for i := range n {
		nu := NewUser{
			Name:       name.MustParse(fmt.Sprintf("Name%d", i)),
			Email:      mail.Address{Address: fmt.Sprintf("Email%d-%s@gmail.com", i, suffix)},
			Department: name.MustParseNull(fmt.Sprintf("Department%d", i)),
			Password:   password.MustParse("Secret123!@#"),
		}

		newUsrs[i] = nu
	}

	return newUsrs
}
