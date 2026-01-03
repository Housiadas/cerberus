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

	tu1 := apitest.User{
		User: usrs[0],
	}

	tu2 := apitest.User{
		User: usrs[1],
	}

	// -------------------------------------------------------------------------

	sd := apitest.SeedData{
		Users: []apitest.User{tu1, tu2},
		//Admins: []apitest.User{tu1, tu2},
	}

	return sd, nil
}
