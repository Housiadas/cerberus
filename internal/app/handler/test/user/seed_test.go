package user_test

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/internal/common/dbtest"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
)

func insertSeedData(db *dbtest.Database) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.Core

	usrs, err := user_service.TestSeedUsers(ctx, 2, busDomain.User)
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
