package permission_repo_test

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	permission2 "github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/core/permission/permission_repo"
	"github.com/Housiadas/cerberus/internal/sdk/eventbus"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/dbtest"
	unitest2 "github.com/Housiadas/cerberus/internal/sdk/testutil/unitest"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func Test_Permission(t *testing.T) {
	t.Parallel()

	db := sc.NewDB(t)

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	uuidGen := uuidgen.NewV7()
	tx := pgsql.NewTransactor(log, db)
	permService := permission2.NewService(log, permission_repo.NewStore(log, db), uuidGen, tx, eventbus.NewNop())

	sd, err := insertSeedData(permService)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest2.Run(t, queryPermission(permService, sd), "query")
	unitest2.Run(t, createPermission(permService), "create")
	unitest2.Run(t, updatePermission(permService, sd), "update")
	unitest2.Run(t, deletePermission(permService, sd), "delete")
}

func insertSeedData(service *permission2.Service) (unitest2.SeedData, error) {
	ctx := context.Background()

	perms, err := permission2.TestSeedPermissions(ctx, 2, service)
	if err != nil {
		return unitest2.SeedData{}, fmt.Errorf("seeding permissions: %w", err)
	}

	sd := unitest2.SeedData{
		Permissions: []unitest2.Permission{
			{Permission: perms[0]},
			{Permission: perms[1]},
		},
	}

	return sd, nil
}

func queryPermission(service *permission2.Service, sd unitest2.SeedData) []unitest2.Table {
	perms := make([]permission2.Permission, 0, len(sd.Permissions))
	for _, p := range sd.Permissions {
		perms = append(perms, p.Permission)
	}

	sort.Slice(perms, func(i, j int) bool {
		return perms[i].ID().String() <= perms[j].ID().String()
	})

	permCmpOpts := []cmp.Option{
		cmp.AllowUnexported(permission2.Permission{}),
		cmpopts.EquateApproxTime(time.Second),
	}

	return []unitest2.Table{
		{
			Name:    "all",
			ExpResp: perms,
			ExcFunc: func(ctx context.Context) any {
				filter := permission2.QueryFilter{
					Name: dbtest.NamePointer("Permission"),
				}

				resp, err := service.Query(ctx, filter, permission2.GetDefaultOrderBy(), mustParseCursor("", "10"))
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]permission2.Permission)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]permission2.Permission)

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
				gotResp, exists := got.(permission2.Permission)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(permission2.Permission)

				return cmp.Diff(gotResp, expResp, permCmpOpts...)
			},
		},
	}
}

func createPermission(service *permission2.Service) []unitest2.Table {
	return []unitest2.Table{
		{
			Name: "basic",
			ExpResp: permission2.New(
				uuid.UUID{},
				name.MustParse("TestPermission"),
				time.Time{},
				time.Time{},
				nil,
			),
			ExcFunc: func(ctx context.Context) any {
				np := permission2.NewPermission{
					Name: name.MustParse("TestPermission"),
				}

				resp, err := service.Create(ctx, np)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(permission2.Permission)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(permission2.Permission)

				expResp = permission2.New(gotResp.ID(), expResp.Name(), gotResp.CreatedAt(), gotResp.UpdatedAt(), gotResp.DeletedAt())

				return cmp.Diff(gotResp, expResp, cmp.AllowUnexported(permission2.Permission{}))
			},
		},
	}
}

func updatePermission(service *permission2.Service, sd unitest2.SeedData) []unitest2.Table {
	newName := name.MustParse("UpdatedPermission")

	return []unitest2.Table{
		{
			Name: "basic",
			ExpResp: permission2.New(
				sd.Permissions[0].ID(),
				newName,
				sd.Permissions[0].CreatedAt(),
				time.Time{},
				nil,
			),
			ExcFunc: func(ctx context.Context) any {
				up := permission2.UpdatePermission{
					Name: &newName,
				}

				resp, err := service.Update(ctx, sd.Permissions[0].Permission, up)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(permission2.Permission)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(permission2.Permission)
				expResp = expResp.WithUpdatedAt(gotResp.UpdatedAt())

				return cmp.Diff(gotResp, expResp, cmp.AllowUnexported(permission2.Permission{}), cmpopts.EquateApproxTime(time.Second))
			},
		},
	}
}

func deletePermission(service *permission2.Service, sd unitest2.SeedData) []unitest2.Table {
	return []unitest2.Table{
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
