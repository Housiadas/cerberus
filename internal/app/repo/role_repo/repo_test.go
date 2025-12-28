package role_repo_test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/common/dbtest"
	"github.com/Housiadas/cerberus/internal/common/unitest"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/pkg/page"
	"github.com/google/go-cmp/cmp"
)

func Test_Role(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_Role")

	sd, err := insertSeedData(db.Core)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, query(db.Core, sd), "query")
	unitest.Run(t, create(db.Core), "create")
	unitest.Run(t, update(db.Core, sd), "update")
	unitest.Run(t, deleteRole(db.Core, sd), "delete")
}

// =============================================================================

func insertSeedData(service dbtest.Service) (unitest.SeedData, error) {
	ctx := context.Background()

	roles, err := role_service.TestSeedRoles(ctx, 2, service.Role)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding roles : %w", err)
	}

	tu1 := unitest.Role{
		Role: roles[0],
	}

	tu2 := unitest.Role{
		Role: roles[1],
	}

	sd := unitest.SeedData{
		Roles: []unitest.Role{tu1, tu2},
	}

	return sd, nil
}

func query(service dbtest.Service, sd unitest.SeedData) []unitest.Table {
	roles := make([]role.Role, 0, len(sd.Roles))

	for _, rls := range sd.Roles {
		roles = append(roles, rls.Role)
	}

	sort.Slice(roles, func(i, j int) bool {
		return roles[i].ID.String() <= roles[j].ID.String()
	})

	table := []unitest.Table{
		{
			Name:    "all",
			ExpResp: roles,
			ExcFunc: func(ctx context.Context) any {
				filter := role.QueryFilter{
					Name: dbtest.NamePointer("Name"),
				}

				resp, err := service.Role.Query(ctx, filter, role.DefaultOrderBy, page.MustParse("1", "10"))
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]role.Role)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]role.Role)

				for i := range gotResp {
					if gotResp[i].CreatedAt.Format(time.RFC3339) == expResp[i].CreatedAt.Format(time.RFC3339) {
						expResp[i].CreatedAt = gotResp[i].CreatedAt
					}

					if gotResp[i].UpdatedAt.Format(time.RFC3339) == expResp[i].UpdatedAt.Format(time.RFC3339) {
						expResp[i].UpdatedAt = gotResp[i].UpdatedAt
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func create(service dbtest.Service) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "basic",
			ExpResp: role.Role{
				Name: name.MustParse("Admin"),
			},
			ExcFunc: func(ctx context.Context) any {
				nr := role.NewRole{
					Name: name.MustParse("Admin"),
				}

				resp, err := service.Role.Create(ctx, nr)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(role.Role)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(role.Role)

				expResp.ID = gotResp.ID
				expResp.CreatedAt = gotResp.CreatedAt
				expResp.UpdatedAt = gotResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func update(service dbtest.Service, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "basic",
			ExpResp: role.Role{
				ID:        sd.Roles[0].ID,
				Name:      name.MustParse("Chris Housi 2"),
				CreatedAt: sd.Roles[0].CreatedAt,
			},
			ExcFunc: func(ctx context.Context) any {
				urole := role.UpdateRole{
					Name: dbtest.NamePointer("Chris Housi 2"),
				}

				resp, err := service.Role.Update(ctx, sd.Roles[0].Role, urole)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(role.Role)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(role.Role)
				expResp.UpdatedAt = gotResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func deleteRole(service dbtest.Service, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "role 1",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := service.Role.Delete(ctx, sd.Roles[1].Role); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
