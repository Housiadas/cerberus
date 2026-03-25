package handlers_test

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	apitest2 "github.com/Housiadas/cerberus/internal/testutil/apitest"
)

func insertPermissionSeedData(test *apitest2.Test) (apitest2.SeedData, error) {
	ctx := context.Background()

	usrs, err := user_service.TestSeedUsers(ctx, 2, test.Core.User)
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding users: %w", err)
	}

	adminRoleID, err := apitest2.SeedRole(ctx, test.DB, "admin")
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding admin role: %w", err)
	}

	for _, permName := range []string{"permission:read", "permission:write", "permission:delete", "user:read", "user:write"} {
		pid, seedErr := apitest2.SeedPermission(ctx, test.DB, permName)
		if seedErr != nil {
			return apitest2.SeedData{}, fmt.Errorf("seeding permission %s: %w", permName, seedErr)
		}

		if seedErr = apitest2.SeedRolePermission(ctx, test.DB, adminRoleID, pid); seedErr != nil {
			return apitest2.SeedData{}, fmt.Errorf("seeding admin role permission %s: %w", permName, seedErr)
		}
	}

	userRoleID, err := apitest2.SeedRole(ctx, test.DB, "user")
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding user role: %w", err)
	}

	for _, permName := range []string{"user:read", "user:write"} {
		pid, seedErr := apitest2.SeedPermission(ctx, test.DB, permName+":user")
		if seedErr != nil {
			return apitest2.SeedData{}, fmt.Errorf("seeding permission %s: %w", permName, seedErr)
		}

		if seedErr = apitest2.SeedRolePermission(ctx, test.DB, userRoleID, pid); seedErr != nil {
			return apitest2.SeedData{}, fmt.Errorf("seeding user role permission %s: %w", permName, seedErr)
		}
	}

	if err = apitest2.SeedUserRole(ctx, test.DB, usrs[0].ID(), adminRoleID); err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding admin user role: %w", err)
	}

	if err = apitest2.SeedUserRole(ctx, test.DB, usrs[1].ID(), userRoleID); err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding regular user role: %w", err)
	}

	tkn1, err := test.Usecase.Auth.GenerateAccessToken(ctx, usrs[0].ID().String())
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding token: %w", err)
	}

	tkn2, err := test.Usecase.Auth.GenerateAccessToken(ctx, usrs[1].ID().String())
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding token: %w", err)
	}

	return apitest2.SeedData{
		Admins: []apitest2.User{
			{User: usrs[0], AccessToken: tkn1},
		},
		Users: []apitest2.User{
			{User: usrs[1], AccessToken: tkn2},
		},
	}, nil
}
