package product_test

import (
	"context"
	"fmt"
	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/internal/core/service/user_core"

	"github.com/Housiadas/cerberus/internal/common/dbtest"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/internal/core/service/product_core"
)

func insertSeedData(db *dbtest.Database) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.Core

	usrs, err := user_core.TestSeedUsers(ctx, 1, role.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err := product_core.TestGenerateSeedProducts(ctx, 2, busDomain.Product, usrs[0].ID)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding products : %w", err)
	}

	tu1 := apitest.User{
		User:     usrs[0],
		Products: prds,
	}

	// -------------------------------------------------------------------------

	usrs, err = user_core.TestSeedUsers(ctx, 1, role.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	prds, err = product_core.TestGenerateSeedProducts(ctx, 2, busDomain.Product, usrs[0].ID)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding products : %w", err)
	}

	tu2 := apitest.User{
		User:     usrs[0],
		Products: prds,
	}

	// -------------------------------------------------------------------------

	sd := apitest.SeedData{
		Admins: []apitest.User{tu2},
		Users:  []apitest.User{tu1},
	}

	return sd, nil
}
