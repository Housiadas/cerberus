package handlers_test

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/user"
	apitest2 "github.com/Housiadas/cerberus/internal/testutil/apitest"
)

func insertAuthSeedData(test *apitest2.Test) (apitest2.SeedData, error) {
	ctx := context.Background()

	usrs, err := user.TestSeedUsers(ctx, 2, test.Core.User)
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tkn1, err := test.Usecase.Auth.GenerateAccessToken(ctx, usrs[0].ID().String())
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding token : %w", err)
	}

	tkn2, err := test.Usecase.Auth.GenerateAccessToken(ctx, usrs[1].ID().String())
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding token : %w", err)
	}

	return apitest2.SeedData{
		Users: []apitest2.User{
			{User: usrs[0], AccessToken: tkn1},
			{User: usrs[1], AccessToken: tkn2},
		},
	}, nil
}
