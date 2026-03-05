package permission_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
)

// TestNewPermissions is a helper method for testing.
func TestNewPermissions(n int) []permission.NewPermission {
	newPerms := make([]permission.NewPermission, n)

	for i := range n {
		np := permission.NewPermission{
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
) ([]permission.Permission, error) {
	newPerms := TestNewPermissions(n)

	perms := make([]permission.Permission, len(newPerms))

	for i, np := range newPerms {
		perm, err := service.Create(ctx, np)
		if err != nil {
			return nil, fmt.Errorf("seeding permission: idx: %d : %w", i, err)
		}

		perms[i] = perm
	}

	return perms, nil
}
