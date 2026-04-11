package handlers_test

import (
	"context"
	"fmt"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/apitest"
	"github.com/google/uuid"
)

func insertUserSeedData(test *apitest.Test) (apitest.SeedData, error) {
	ctx := context.Background()

	store := db.NewStore(test.DB)

	usrs, err := user.TestSeedUsers(ctx, 2, test.Core.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	permIDs := make(map[string]uuid.UUID)
	for _, name := range []string{"user:read:all", "user:delete", "user:read", "user:write"} {
		id, err := apitest.SeedPermission(ctx, store, name)
		if err != nil {
			return apitest.SeedData{}, fmt.Errorf("seeding permission %s : %w", name, err)
		}
		permIDs[name] = id
	}

	adminRoleID, err := apitest.SeedRole(ctx, store, "admin")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin role : %w", err)
	}

	for _, pid := range permIDs {
		if err = apitest.SeedRolePermission(ctx, store, adminRoleID, pid); err != nil {
			return apitest.SeedData{}, fmt.Errorf("seeding admin role permission : %w", err)
		}
	}

	userRoleID, err := apitest.SeedRole(ctx, store, "user")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user role : %w", err)
	}

	for _, name := range []string{"user:read", "user:write"} {
		if err = apitest.SeedRolePermission(ctx, store, userRoleID, permIDs[name]); err != nil {
			return apitest.SeedData{}, fmt.Errorf("seeding user role permission : %w", err)
		}
	}

	if err = apitest.SeedUserRole(ctx, store, usrs[0].ID(), adminRoleID); err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin user role : %w", err)
	}

	if err = apitest.SeedUserRole(ctx, store, usrs[1].ID(), userRoleID); err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding regular user role : %w", err)
	}

	tkn1, err := test.Core.Auth.GenerateAccessToken(ctx, usrs[0].ID().String())
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding token : %w", err)
	}

	tkn2, err := test.Core.Auth.GenerateAccessToken(ctx, usrs[1].ID().String())
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding token : %w", err)
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
