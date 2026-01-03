package audit_test

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/internal/core/domain/entity"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
)

func insertSeedData(test *apitest.Test) (apitest.SeedData, error) {
	ctx := context.Background()
	usrs, err := user_service.TestSeedUsers(ctx, 1, test.Core.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	audits, err := audit_service.TestSeedAudits(
		ctx, 2, usrs[0].ID, entity.User, "create", test.Core.Audit,
	)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := apitest.User{
		User:   usrs[0],
		Audits: audits,
	}

	// -------------------------------------------------------------------------

	sd := apitest.SeedData{
		Admins: []apitest.User{tu1},
	}

	return sd, nil
}
