package permission_repo_test

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/core/permission/permission_repo"
	"github.com/Housiadas/cerberus/internal/sdk/eventbus"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/dbtest"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/unitest"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

func Test_Permission(t *testing.T) {
	t.Parallel()

	db := sc.NewDB(t)

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	clk := clock.NewClock()
	uuidGen := uuidgen.NewV7()
	tx := pgsql.NewTransactor(log, db)
	permService := permission.NewService(log, permission_repo.NewStore(log, db), uuidGen, tx, eventbus.NewNop(), clk)

	sd, err := insertSeedData(permService)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, queryPermission(permService, sd), "query")
	unitest.Run(t, createPermission(permService), "create")
	unitest.Run(t, updatePermission(permService, sd), "update")
	unitest.Run(t, deletePermission(permService, sd), "delete")
}

func insertSeedData(service *permission.Service) (unitest.SeedData, error) {
	ctx := context.Background()

	perms, err := permission.TestSeedPermissions(ctx, 2, service)
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

func queryPermission(service *permission.Service, sd unitest.SeedData) []unitest.Table {
	perms := make([]permission.Permission, 0, len(sd.Permissions))
	for _, p := range sd.Permissions {
		perms = append(perms, p.Permission)
	}

	sort.Slice(perms, func(i, j int) bool {
		return perms[i].ID().String() <= perms[j].ID().String()
	})

	permCmpOpts := []cmp.Option{
		cmp.AllowUnexported(permission.Permission{}),
		cmpopts.EquateApproxTime(time.Second),
	}

	return []unitest.Table{
		{
			Name:    "all",
			ExpResp: perms,
			ExcFunc: func(ctx context.Context) any {
				filter := permission.QueryFilter{
					Name: dbtest.NamePointer("Permission"),
				}

				resp, err := service.Query(ctx, filter, permission.GetDefaultOrderBy(), mustParseCursor("", "10"))
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

				return cmp.Diff(gotResp, expResp, permCmpOpts...)
			},
		},
		{
			Name:    "byid",
			ExpResp: sd.Permissions[0].Permission,
			ExcFunc: func(ctx context.Context) any {
				resp, err := service.QueryByID(ctx, sd.Permissions[0].ID())
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

				return cmp.Diff(gotResp, expResp, permCmpOpts...)
			},
		},
	}
}

func createPermission(service *permission.Service) []unitest.Table {
	return []unitest.Table{
		{
			Name: "basic",
			ExpResp: permission.New(
				uuid.UUID{},
				name.MustParse("TestPermission"),
				time.Time{},
				time.Time{},
				nil,
			),
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

				expResp = permission.New(gotResp.ID(), expResp.Name(), gotResp.CreatedAt(), gotResp.UpdatedAt(), gotResp.DeletedAt())

				return cmp.Diff(gotResp, expResp, cmp.AllowUnexported(permission.Permission{}))
			},
		},
	}
}

func updatePermission(service *permission.Service, sd unitest.SeedData) []unitest.Table {
	newName := name.MustParse("UpdatedPermission")

	return []unitest.Table{
		{
			Name: "basic",
			ExpResp: permission.New(
				sd.Permissions[0].ID(),
				newName,
				sd.Permissions[0].CreatedAt(),
				time.Time{},
				nil,
			),
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
				expResp = expResp.WithUpdatedAt(gotResp.UpdatedAt())

				return cmp.Diff(gotResp, expResp, cmp.AllowUnexported(permission.Permission{}), cmpopts.EquateApproxTime(time.Second))
			},
		},
	}
}

func deletePermission(service *permission.Service, sd unitest.SeedData) []unitest.Table {
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

func mustParseCursor(cursorStr, limitStr string) cursor.Cursor {
	cur, err := cursor.Parse(cursorStr, limitStr)
	if err != nil {
		panic(err)
	}

	return cur
}
