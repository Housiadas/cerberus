package role

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/types/name"
)

// TestNewRoles is a helper method for testing.
func TestNewRoles(n int) []NewRole {
	newRoles := make([]NewRole, n)

	for i := range n {
		nrole := NewRole{
			Name: name.MustParse(fmt.Sprintf("Name%d", i)),
		}

		newRoles[i] = nrole
	}

	return newRoles
}

// TestSeedRoles is a helper method for testing.
func TestSeedRoles(ctx context.Context, n int, service *Service) ([]Role, error) {
	newRoles := TestNewRoles(n)

	roles := make([]Role, len(newRoles))

	for i, nu := range newRoles {
		nrole, err := service.Create(ctx, nu)
		if err != nil {
			return nil, fmt.Errorf("seeding role: idx: %d : %w", i, err)
		}

		roles[i] = nrole
	}

	return roles, nil
}
