package handlers_test

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/entity"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/internal/utils/apitest"
)

func insertAuditSeedData(test *apitest.Test) (apitest.SeedData, error) {
	ctx := context.Background()

	usrs, err := user_service.TestSeedUsers(ctx, 2, test.Core.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	audits, err := audit_service.TestSeedAudits(
		ctx, 2, usrs[0].ID(), entity.New(entity.UserEntity), "create", test.Core.Audit,
	)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding audits : %w", err)
	}

	auditReadID, err := apitest.SeedPermission(ctx, test.DB, "audit:read")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding permission : %w", err)
	}

	adminRoleID, err := apitest.SeedRole(ctx, test.DB, "admin")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin role : %w", err)
	}

	if err = apitest.SeedRolePermission(ctx, test.DB, adminRoleID, auditReadID); err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin role permission : %w", err)
	}

	userRoleID, err := apitest.SeedRole(ctx, test.DB, "user")
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user role : %w", err)
	}

	if err = apitest.SeedUserRole(ctx, test.DB, usrs[0].ID(), adminRoleID); err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admin user role : %w", err)
	}

	if err = apitest.SeedUserRole(ctx, test.DB, usrs[1].ID(), userRoleID); err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding regular user role : %w", err)
	}

	tkn1, err := test.Usecase.Auth.GenerateAccessToken(ctx, usrs[0].ID().String())
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding token : %w", err)
	}

	tkn2, err := test.Usecase.Auth.GenerateAccessToken(ctx, usrs[1].ID().String())
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
