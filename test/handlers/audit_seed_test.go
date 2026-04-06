package handlers_test

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/user"
	apitest2 "github.com/Housiadas/cerberus/internal/testutil/apitest"
	"github.com/Housiadas/cerberus/internal/types/entity"
)

func insertAuditSeedData(test *apitest2.Test) (apitest2.SeedData, error) {
	ctx := context.Background()

	usrs, err := user.TestSeedUsers(ctx, 2, test.Core.User)
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	audits, err := audit.TestSeedAudits(
		ctx, 2, usrs[0].ID(), entity.New(entity.UserEntity), "create", test.Core.Audit,
	)
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding audits : %w", err)
	}

	auditReadID, err := apitest2.SeedPermission(ctx, test.DB, "audit:read")
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding permission : %w", err)
	}

	adminRoleID, err := apitest2.SeedRole(ctx, test.DB, "admin")
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding admin role : %w", err)
	}

	if err = apitest2.SeedRolePermission(ctx, test.DB, adminRoleID, auditReadID); err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding admin role permission : %w", err)
	}

	userRoleID, err := apitest2.SeedRole(ctx, test.DB, "user")
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding user role : %w", err)
	}

	if err = apitest2.SeedUserRole(ctx, test.DB, usrs[0].ID(), adminRoleID); err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding admin user role : %w", err)
	}

	if err = apitest2.SeedUserRole(ctx, test.DB, usrs[1].ID(), userRoleID); err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding regular user role : %w", err)
	}

	tkn1, err := test.Auth.GenerateAccessToken(ctx, usrs[0].ID().String())
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding token : %w", err)
	}

	tkn2, err := test.Auth.GenerateAccessToken(ctx, usrs[1].ID().String())
	if err != nil {
		return apitest2.SeedData{}, fmt.Errorf("seeding token : %w", err)
	}

	return apitest2.SeedData{
		Admins: []apitest2.User{
			{User: usrs[0], AccessToken: tkn1, Audits: audits},
		},
		Users: []apitest2.User{
			{User: usrs[1], AccessToken: tkn2},
		},
	}, nil
}
