package handlers_test

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/apitest"
)

func insertAuthSeedData(test *apitest.Test) (apitest.SeedData, error) {
	ctx := context.Background()

	usrs, err := user.TestSeedUsers(ctx, 2, test.Core.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tkn1, err := test.Auth.GenerateAccessToken(ctx, usrs[0].ID().String())
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding token : %w", err)
	}

	tkn2, err := test.Auth.GenerateAccessToken(ctx, usrs[1].ID().String())
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding token : %w", err)
	}

	return apitest.SeedData{
		Users: []apitest.User{
			{User: usrs[0], AccessToken: tkn1},
			{User: usrs[1], AccessToken: tkn2},
		},
	}, nil
}
