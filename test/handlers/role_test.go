package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/Housiadas/cerberus/internal/core/role"
	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/testutil/apitest"
	"github.com/Housiadas/cerberus/internal/testutil/dbtest"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/pkg/clock"
)

func Test_API_Role_Query_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertRoleSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	roles, err := role.TestSeedRoles(ctx, 2, test.Core.Role)
	require.NoError(t, err)

	sort.Slice(roles, func(i, j int) bool {
		return roles[i].ID().String() <= roles[j].ID().String()
	})

	expData := toTestRoles(roles)
	expMd := openapi.Metadata{
		HasMore: false,
		Limit:   10,
	}

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         "/api/v1/roles?limit=10&orderBy=id,ASC&name=Name",
			Method:      http.MethodGet,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			GotResp:     &openapi.RolePageResult{},
			ExpResp: &openapi.RolePageResult{
				Data:     &expData,
				Metadata: &expMd,
			},
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "role-query-200")
}

func Test_API_Role_Query_403(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertRoleSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "role-list-forbidden",
			URL:         "/api/v1/roles?limit=10",
			Method:      http.MethodGet,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			GotResp:     &errs2.Error{},
			ExpResp:     errs2.Errorf(errs2.PermissionDenied, errs2.CodePermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "role-query-403")
}

func Test_API_Role_Create_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertRoleSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         "/api/v1/roles",
			Method:      http.MethodPost,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Input: &openapi.NewRole{
				Name: "editor",
			},
			GotResp: &openapi.Role{},
			ExpResp: &openapi.Role{
				Name: "editor",
			},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*openapi.Role)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*openapi.Role)
				expResp.Id = gotResp.Id
				expResp.CreatedAt = gotResp.CreatedAt
				expResp.UpdatedAt = gotResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	test.Run(t, table, "role-create-200")
}

func Test_API_Role_Create_400(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertRoleSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "missing-name",
			URL:         "/api/v1/roles",
			Method:      http.MethodPost,
			StatusCode:  http.StatusBadRequest,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Input:       &openapi.NewRole{},
			GotResp:     &errs2.Error{},
			ExpResp: &errs2.Error{
				Status:  errs2.InvalidArgument,
				Code:    errs2.CodeValidation,
				Message: "validation error",
				Fields:  []errs2.FieldError{{Field: "name", Err: "name is required"}},
			},
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-access-token",
			URL:        "/api/v1/roles",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input:      &openapi.NewRole{Name: "editor"},
			GotResp:    &errs2.Error{},
			ExpResp:    errs2.Errorf(errs2.Unauthenticated, errs2.CodeUnauthenticated, "expected authorization header format: Bearer <token>"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "role-create-400")
}

func Test_API_Role_Create_403(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertRoleSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         "/api/v1/roles",
			Method:      http.MethodPost,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input:       &openapi.NewRole{Name: "editor"},
			GotResp:     &errs2.Error{},
			ExpResp:     errs2.Errorf(errs2.PermissionDenied, errs2.CodePermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "role-create-403")
}

func Test_API_Role_Update_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertRoleSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	roles, err := role.TestSeedRoles(ctx, 1, test.Core.Role)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         fmt.Sprintf("/api/v1/roles/%s", roles[0].ID()),
			Method:      http.MethodPut,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Input: &openapi.UpdateRole{
				Name: dbtest.StringPointer("UpdatedRole"),
			},
			GotResp: &openapi.Role{},
			ExpResp: func() *openapi.Role {
				return &openapi.Role{
					Id:        roles[0].ID().String(),
					Name:      "UpdatedRole",
					CreatedAt: clock.Format(new(roles[0].CreatedAt())),
					UpdatedAt: clock.Format(new(roles[0].UpdatedAt())),
				}
			}(),
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*openapi.Role)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*openapi.Role)
				gotResp.UpdatedAt = expResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	test.Run(t, table, "role-update-200")
}

func Test_API_Role_Update_403(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertRoleSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	roles, err := role.TestSeedRoles(ctx, 1, test.Core.Role)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         fmt.Sprintf("/api/v1/roles/%s", roles[0].ID()),
			Method:      http.MethodPut,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &openapi.UpdateRole{
				Name: dbtest.StringPointer("UpdatedRole"),
			},
			GotResp: &errs2.Error{},
			ExpResp: errs2.Errorf(errs2.PermissionDenied, errs2.CodePermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "role-update-403")
}

func Test_API_Role_Delete_204(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertRoleSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	roles, err := role.TestSeedRoles(ctx, 1, test.Core.Role)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         fmt.Sprintf("/api/v1/roles/%s", roles[0].ID()),
			Method:      http.MethodDelete,
			StatusCode:  http.StatusNoContent,
			AccessToken: &sd.Admins[0].AccessToken.Token,
		},
	}

	test.Run(t, table, "role-delete-204")
}

func Test_API_Role_Delete_403(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertRoleSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	roles, err := role.TestSeedRoles(ctx, 1, test.Core.Role)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         fmt.Sprintf("/api/v1/roles/%s", roles[0].ID()),
			Method:      http.MethodDelete,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			GotResp:     &errs2.Error{},
			ExpResp:     errs2.Errorf(errs2.PermissionDenied, errs2.CodePermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "role-delete-403")
}
