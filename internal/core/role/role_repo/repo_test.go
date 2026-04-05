package role_repo_test

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/core/role/role_repo"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/role"
	"github.com/Housiadas/cerberus/internal/eventbus"
	"github.com/Housiadas/cerberus/internal/testutil/dbtest"
	"github.com/Housiadas/cerberus/internal/testutil/unitest"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

func Test_Role(t *testing.T) {
	t.Parallel()

	db := sc.NewDB(t)

	var buf bytes.Buffer
	traceIDFn := func(context.Context) string { return "" }
	requestIDFn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	uuidGen := uuidgen.NewV7()
	tx := pgsql.NewTransactor(log, db)
	roleService := role.NewService(log, role_repo.NewStore(log, db), uuidGen, tx, eventbus.NewNop())

	sd, err := insertSeedData(roleService)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, queryRole(roleService, sd), "query")
	unitest.Run(t, createRole(roleService), "create")
	unitest.Run(t, updateRole(roleService, sd), "update")
	unitest.Run(t, deleteRole(roleService, sd), "delete")
}

func insertSeedData(service *role.Service) (unitest.SeedData, error) {
	ctx := context.Background()

	roles, err := role.TestSeedRoles(ctx, 2, service)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	sd := unitest.SeedData{
		Roles: []unitest.Role{
			{Role: roles[0]},
			{Role: roles[1]},
		},
	}

	return sd, nil
}

func queryRole(service *role.Service, sd unitest.SeedData) []unitest.Table {
	roles := make([]role.Role, 0, len(sd.Roles))
	for _, r := range sd.Roles {
		roles = append(roles, r.Role)
	}

	sort.Slice(roles, func(i, j int) bool {
		return roles[i].ID().String() <= roles[j].ID().String()
	})

	roleCmpOpts := []cmp.Option{
		cmp.AllowUnexported(role.Role{}),
		cmpopts.EquateApproxTime(time.Second),
	}

	return []unitest.Table{
		{
			Name:    "all",
			ExpResp: roles,
			ExcFunc: func(ctx context.Context) any {
				filter := role.QueryFilter{
					Name: dbtest.NamePointer("Name"),
				}

				resp, err := service.Query(ctx, filter, role.GetDefaultOrderBy(), mustParseCursor("", "10"))
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

				return cmp.Diff(gotResp, expResp, roleCmpOpts...)
			},
		},
		{
			Name:    "byid",
			ExpResp: sd.Roles[0].Role,
			ExcFunc: func(ctx context.Context) any {
				resp, err := service.QueryByID(ctx, sd.Roles[0].ID())
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

				return cmp.Diff(gotResp, expResp, roleCmpOpts...)
			},
		},
	}
}

func createRole(service *role.Service) []unitest.Table {
	return []unitest.Table{
		{
			Name: "basic",
			ExpResp: role.New(
				uuid.UUID{},
				name.MustParse("TestRole"),
				time.Time{},
				time.Time{},
				nil,
			),
			ExcFunc: func(ctx context.Context) any {
				nr := role.NewRole{
					Name: name.MustParse("TestRole"),
				}

				resp, err := service.Create(ctx, nr)
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

				expResp = role.New(gotResp.ID(), expResp.Name(), gotResp.CreatedAt(), gotResp.UpdatedAt(), gotResp.DeletedAt())

				return cmp.Diff(gotResp, expResp, cmp.AllowUnexported(role.Role{}))
			},
		},
	}
}

func updateRole(service *role.Service, sd unitest.SeedData) []unitest.Table {
	newName := name.MustParse("UpdatedRole")

	return []unitest.Table{
		{
			Name: "basic",
			ExpResp: role.New(
				sd.Roles[0].ID(),
				newName,
				sd.Roles[0].CreatedAt(),
				time.Time{},
				nil,
			),
			ExcFunc: func(ctx context.Context) any {
				ur := role.UpdateRole{
					Name: &newName,
				}

				resp, err := service.Update(ctx, sd.Roles[0].Role, ur)
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
				expResp = expResp.WithUpdatedAt(gotResp.UpdatedAt())

				return cmp.Diff(gotResp, expResp, cmp.AllowUnexported(role.Role{}), cmpopts.EquateApproxTime(time.Second))
			},
		},
	}
}

func deleteRole(service *role.Service, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "role",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := service.Delete(ctx, sd.Roles[1].Role); err != nil {
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
