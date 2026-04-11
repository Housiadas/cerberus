package handlers_test

import (
	"context"
	"fmt"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/apitest"
	"github.com/Housiadas/cerberus/internal/types/entity"
)

func insertAuditSeedData(test *apitest.Test) (apitest.SeedData, error) {
	ctx := context.Background()

	store := db.NewStore(test.DB)

	usrs, err := user.TestSeedUsers(ctx, 2, test.Core.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	audits, err := audit.TestSeedAudits(
		ctx, 2, usrs[0].ID(), entity.New(entity.UserEntity), "create", test.Core.Audit,
	)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding audits : %w", err)
	}

	auditReadID, err := apitest.SeedPermission(ctx, store, "audit:read")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding permission : %w", err)
	}

	adminRoleID, err := apitest.SeedRole(ctx, store, "admin")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin role : %w", err)
	}

	if err = apitest.SeedRolePermission(ctx, store, adminRoleID, auditReadID); err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin role permission : %w", err)
	}

	userRoleID, err := apitest.SeedRole(ctx, store, "user")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user role : %w", err)
	}

	if err = apitest.SeedUserRole(ctx, store, usrs[0].ID(), adminRoleID); err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin user role : %w", err)
	}

	if err = apitest.SeedUserRole(ctx, store, usrs[1].ID(), userRoleID); err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding regular user role : %w", err)
	}

	tkn1, err := test.Core.Auth.GenerateAccessToken(ctx, usrs[0].ID().String())
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding token : %w", err)
	}

	tkn2, err := test.Core.Auth.GenerateAccessToken(ctx, usrs[1].ID().String())
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding token : %w", err)
	}

	return apitest.SeedData{
		Admins: []apitest.User{
			{User: usrs[0], AccessToken: tkn1, Audits: audits},
		},
		Users: []apitest.User{
			{User: usrs[1], AccessToken: tkn2},
		},
	}, nil
}
