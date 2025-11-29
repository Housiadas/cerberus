package audit_test

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/internal/common/dbtest"
	"github.com/Housiadas/cerberus/internal/core/domain/entity"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/internal/core/service/audit_core"
	"github.com/Housiadas/cerberus/internal/core/service/user_core"
)

func insertSeedData(db *dbtest.Database) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.Core

	usrs, err := user_core.TestSeedUsers(ctx, 1, role.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	audits, err := audit_core.TestSeedAudits(ctx, 2, usrs[0].ID, entity.User, "create", busDomain.Audit)
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
