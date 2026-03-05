package permission_repo_test

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/Housiadas/cerberus/internal/app/repo/permission_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/internal/utils/dbtest"
	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/Housiadas/cerberus/internal/utils/unitest"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"

	"github.com/google/go-cmp/cmp"
)

func Test_Permission(t *testing.T) {
	t.Parallel()

	db := sc.NewDB(t)

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	uuidGen := uuidgen.NewV7()
	permService := permission_service.New(log, permission_repo.NewStore(log, db), uuidGen)

	sd, err := insertSeedData(permService)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, queryPermission(permService, sd), "query")
	unitest.Run(t, createPermission(permService), "create")
	unitest.Run(t, updatePermission(permService, sd), "update")
	unitest.Run(t, deletePermission(permService, sd), "delete")
}

func insertSeedData(service *permission_service.Service) (unitest.SeedData, error) {
	ctx := context.Background()

	perms, err := permission_service.TestSeedPermissions(ctx, 2, service)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding permissions: %w", err)
	}

	sd := unitest.SeedData{
		Permissions: []unitest.Permission{
			{Permission: perms[0]},
			{Permission: perms[1]},
		},
	}

	return sd, nil
}

func queryPermission(service *permission_service.Service, sd unitest.SeedData) []unitest.Table {
	perms := make([]permission.Permission, 0, len(sd.Permissions))
	for _, p := range sd.Permissions {
		perms = append(perms, p.Permission)
	}

	sort.Slice(perms, func(i, j int) bool {
		return perms[i].ID.String() <= perms[j].ID.String()
	})

	return []unitest.Table{
		{
			Name:    "all",
			ExpResp: perms,
			ExcFunc: func(ctx context.Context) any {
				filter := permission.QueryFilter{
					Name: dbtest.NamePointer("Permission"),
				}

				resp, err := service.Query(ctx, filter, permission.GetDefaultOrderBy(), page.MustParse("1", "10"))
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]permission.Permission)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]permission.Permission)

				for i := range gotResp {
					if clock.Format(&gotResp[i].CreatedAt) == clock.Format(&expResp[i].CreatedAt) {
						expResp[i].CreatedAt = gotResp[i].CreatedAt
					}

					if clock.Format(&gotResp[i].UpdatedAt) == clock.Format(&expResp[i].UpdatedAt) {
						expResp[i].UpdatedAt = gotResp[i].UpdatedAt
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:    "byid",
			ExpResp: sd.Permissions[0].Permission,
			ExcFunc: func(ctx context.Context) any {
				resp, err := service.QueryByID(ctx, sd.Permissions[0].ID)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(permission.Permission)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(permission.Permission)

				if clock.Format(&gotResp.CreatedAt) == clock.Format(&expResp.CreatedAt) {
					expResp.CreatedAt = gotResp.CreatedAt
				}

				if clock.Format(&gotResp.UpdatedAt) == clock.Format(&expResp.UpdatedAt) {
					expResp.UpdatedAt = gotResp.UpdatedAt
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func createPermission(service *permission_service.Service) []unitest.Table {
	return []unitest.Table{
		{
			Name: "basic",
			ExpResp: permission.Permission{
				Name: name.MustParse("TestPermission"),
			},
			ExcFunc: func(ctx context.Context) any {
				np := permission.NewPermission{
					Name: name.MustParse("TestPermission"),
				}

				resp, err := service.Create(ctx, np)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(permission.Permission)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(permission.Permission)

				expResp.ID = gotResp.ID
				expResp.CreatedAt = gotResp.CreatedAt
				expResp.UpdatedAt = gotResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func updatePermission(service *permission_service.Service, sd unitest.SeedData) []unitest.Table {
	newName := name.MustParse("UpdatedPermission")

	return []unitest.Table{
		{
			Name: "basic",
			ExpResp: permission.Permission{
				ID:        sd.Permissions[0].ID,
				Name:      newName,
				CreatedAt: sd.Permissions[0].CreatedAt,
			},
			ExcFunc: func(ctx context.Context) any {
				up := permission.UpdatePermission{
					Name: &newName,
				}

				resp, err := service.Update(ctx, sd.Permissions[0].Permission, up)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(permission.Permission)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(permission.Permission)
				expResp.UpdatedAt = gotResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func deletePermission(service *permission_service.Service, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "permission",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := service.Delete(ctx, sd.Permissions[1].Permission); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
