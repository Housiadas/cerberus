package user_test

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
)

func insertSeedData(test *apitest.Test) (apitest.SeedData, error) {
	ctx := context.Background()
	usrs, err := user_service.TestSeedUsers(ctx, 2, test.Core.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	usr1 := usrs[0]
	tkn1, err := test.Usecase.Auth.GenerateAccessToken(ctx, usr1.ID.String())
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding token : %w", err)
	}
	tu1 := apitest.User{
		User:        usr1,
		AccessToken: tkn1,
	}

	usr2 := usrs[1]
	tkn2, err := test.Usecase.Auth.GenerateAccessToken(ctx, usr2.ID.String())
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding token : %w", err)
	}
	tu2 := apitest.User{
		User:        usr2,
		AccessToken: tkn2,
	}

	return apitest.SeedData{
		Users: []apitest.User{tu1, tu2},
	}, nil
}
