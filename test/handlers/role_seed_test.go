package handlers_test

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/apitest"
)

func insertRoleSeedData(test *apitest.Test) (apitest.SeedData, error) {
	ctx := context.Background()

	usrs, err := user.TestSeedUsers(ctx, 2, test.Core.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users: %w", err)
	}

	adminRoleID, err := apitest.SeedRole(ctx, test.DB, "admin")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin role: %w", err)
	}

	for _, permName := range []string{"role:read:all", "role:write", "role:delete", "user:read", "user:write"} {
		pid, seedErr := apitest.SeedPermission(ctx, test.DB, permName)
		if seedErr != nil {
			return apitest.SeedData{}, fmt.Errorf("seeding permission %s: %w", permName, seedErr)
		}

		if seedErr = apitest.SeedRolePermission(ctx, test.DB, adminRoleID, pid); seedErr != nil {
			return apitest.SeedData{}, fmt.Errorf("seeding admin role permission %s: %w", permName, seedErr)
		}
	}

	userRoleID, err := apitest.SeedRole(ctx, test.DB, "user")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user role: %w", err)
	}

	for _, permName := range []string{"user:read:user", "user:write:user"} {
		pid, seedErr := apitest.SeedPermission(ctx, test.DB, permName)
		if seedErr != nil {
			return apitest.SeedData{}, fmt.Errorf("seeding permission %s: %w", permName, seedErr)
		}

		if seedErr = apitest.SeedRolePermission(ctx, test.DB, userRoleID, pid); seedErr != nil {
			return apitest.SeedData{}, fmt.Errorf("seeding user role permission %s: %w", permName, seedErr)
		}
	}

	if err = apitest.SeedUserRole(ctx, test.DB, usrs[0].ID(), adminRoleID); err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin user role: %w", err)
	}

	if err = apitest.SeedUserRole(ctx, test.DB, usrs[1].ID(), userRoleID); err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding regular user role: %w", err)
	}

	tkn1, err := test.Auth.GenerateAccessToken(ctx, usrs[0].ID().String())
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding token: %w", err)
	}

	tkn2, err := test.Auth.GenerateAccessToken(ctx, usrs[1].ID().String())
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding token: %w", err)
	}

	return apitest.SeedData{
		Admins: []apitest.User{
			{User: usrs[0], AccessToken: tkn1},
		},
		Users: []apitest.User{
			{User: usrs[1], AccessToken: tkn2},
		},
	}, nil
}
