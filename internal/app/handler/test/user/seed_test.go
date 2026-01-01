package user_test

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/internal/common/dbtest"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
)

func insertSeedData(db *dbtest.Database) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.Core

	usrs, err := user_service.TestSeedUsers(ctx, 2, role.Admin, busDomain.User)
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

	usrs, err = user_service.TestSeedUsers(ctx, 3, role.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu3 := apitest.User{
		User: usrs[0],
	}

	tu4 := apitest.User{
		User: usrs[1],
	}

	tu5 := apitest.User{
		User: usrs[2],
	}

	// -------------------------------------------------------------------------

	sd := apitest.SeedData{
		Users:  []apitest.User{tu3, tu4, tu5},
		Admins: []apitest.User{tu1, tu2},
	}

	return sd, nil
}
