package role_service

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
)

// TestNewRoles is a helper method for testing.
func TestNewRoles(n int) []role.NewRole {
	newRoles := make([]role.NewRole, n)

	idx := rand.Intn(10000)
	for i := range n {
		idx++
		nrole := role.NewRole{
			Name: name.MustParse(fmt.Sprintf("Name%d", idx)),
		}

		newRoles[i] = nrole
	}

	return newRoles
}

// TestSeedRoles is a helper method for testing.
func TestSeedRoles(ctx context.Context, n int, service *Service) ([]role.Role, error) {
	newRoles := TestNewRoles(n)

	roles := make([]role.Role, len(newRoles))

	for i, nu := range newRoles {
		nrole, err := service.Create(ctx, nu)
		if err != nil {
			return nil, fmt.Errorf("seeding role: idx: %d : %w", i, err)
		}

		roles[i] = nrole
	}

	return roles, nil
}
