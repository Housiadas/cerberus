package user_service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"net/mail"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
)

// TestNewUsers is a helper method for testing.
func TestNewUsers(n int, roleID uuid.UUID) []user.NewUser {
	newUsrs := make([]user.NewUser, n)

	idx := rand.Intn(10000)
	for i := range n {
		idx++

		nu := user.NewUser{
			RoleID:     roleID,
			Name:       name.MustParse(fmt.Sprintf("Name%d", idx)),
			Email:      mail.Address{Address: fmt.Sprintf("Email%d@gmail.com", idx)},
			Department: name.MustParseNull(fmt.Sprintf("Department%d", idx)),
			Password:   fmt.Sprintf("Password%d", idx),
		}

		newUsrs[i] = nu
	}

	return newUsrs
}

// TestSeedUsers is a helper method for testing.
func TestSeedUsers(ctx context.Context, n int, roleID uuid.UUID, service *Service) ([]user.User, error) {
	newUsrs := TestNewUsers(n, roleID)

	usrs := make([]user.User, len(newUsrs))
	for i, nu := range newUsrs {
		usr, err := service.Create(ctx, nu)
		if err != nil {
			return nil, fmt.Errorf("seeding user: idx: %d : %w", i, err)
		}

		usrs[i] = usr
	}

	return usrs, nil
}
