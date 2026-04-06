package permission

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/types/name"
)

// TestNewPermissions is a helper method for testing.
func TestNewPermissions(n int) []NewPermission {
	newPerms := make([]NewPermission, n)

	for i := range n {
		np := NewPermission{
			Name: name.MustParse(fmt.Sprintf("Permission%d", i)),
		}

		newPerms[i] = np
	}

	return newPerms
}

// TestSeedPermissions is a helper method for testing.
func TestSeedPermissions(
	ctx context.Context,
	n int,
	service *Service,
) ([]Permission, error) {
	newPerms := TestNewPermissions(n)

	perms := make([]Permission, len(newPerms))

	for i, np := range newPerms {
		perm, err := service.Create(ctx, np)
		if err != nil {
			return nil, fmt.Errorf("seeding permission: idx: %d : %w", i, err)
		}

		perms[i] = perm
	}

	return perms, nil
}
