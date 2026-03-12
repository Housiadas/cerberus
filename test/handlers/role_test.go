package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/internal/usecase/role_usecase"
	"github.com/Housiadas/cerberus/internal/utils/apitest"
	"github.com/Housiadas/cerberus/internal/utils/dbtest"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/Housiadas/cerberus/pkg/clock"
)

func Test_API_Role_Query_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertRoleSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	roles, err := role_service.TestSeedRoles(ctx, 2, test.Core.Role)
	require.NoError(t, err)

	sort.Slice(roles, func(i, j int) bool {
		return roles[i].ID().String() <= roles[j].ID().String()
	})

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         "/api/v1/roles?page=1&rows=10&orderBy=id,ASC&name=Name",
			Method:      http.MethodGet,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			GotResp:     &page.Result[role_usecase.Role]{},
			ExpResp: &page.Result[role_usecase.Role]{
				Data: toAppRoles(roles),
				Metadata: page.Metadata{
					FirstPage:   1,
					CurrentPage: 1,
					LastPage:    1,
					RowsPerPage: 10,
					Total:       len(roles),
				},
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
			URL:         "/api/v1/roles?page=1&rows=10",
			Method:      http.MethodGet,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			GotResp:     &errs.Error{},
			ExpResp:     errs.Errorf(errs.PermissionDenied, errs.CodePermissionDenied, "permission denied"),
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
			Input: &role_usecase.NewRole{
				Name: "editor",
			},
			GotResp: &role_usecase.Role{},
			ExpResp: &role_usecase.Role{
				Name: "editor",
			},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*role_usecase.Role)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*role_usecase.Role)
				expResp.ID = gotResp.ID
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
			Input:       &role_usecase.NewRole{},
			GotResp:     &errs.Error{},
			ExpResp:     errs.Errorf(errs.InvalidArgument, errs.CodeValidation, "parse: invalid name value: \"\""),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-access-token",
			URL:        "/api/v1/roles",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input:      &role_usecase.NewRole{Name: "editor"},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Errorf(errs.Unauthenticated, errs.CodeUnauthenticated, "expected authorization header format: Bearer <token>"),
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
			Input:       &role_usecase.NewRole{Name: "editor"},
			GotResp:     &errs.Error{},
			ExpResp:     errs.Errorf(errs.PermissionDenied, errs.CodePermissionDenied, "permission denied"),
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
	roles, err := role_service.TestSeedRoles(ctx, 1, test.Core.Role)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         fmt.Sprintf("/api/v1/roles/%s", roles[0].ID()),
			Method:      http.MethodPut,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Input: &role_usecase.UpdateRole{
				Name: dbtest.StringPointer("UpdatedRole"),
			},
			GotResp: &role_usecase.Role{},
			ExpResp: func() *role_usecase.Role {
				createdAt := roles[0].CreatedAt()
				updatedAt := roles[0].UpdatedAt()
				return &role_usecase.Role{
					ID:        roles[0].ID().String(),
					Name:      "UpdatedRole",
					CreatedAt: clock.Format(&createdAt),
					UpdatedAt: clock.Format(&updatedAt),
				}
			}(),
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*role_usecase.Role)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*role_usecase.Role)
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
	roles, err := role_service.TestSeedRoles(ctx, 1, test.Core.Role)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         fmt.Sprintf("/api/v1/roles/%s", roles[0].ID()),
			Method:      http.MethodPut,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &role_usecase.UpdateRole{
				Name: dbtest.StringPointer("UpdatedRole"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Errorf(errs.PermissionDenied, errs.CodePermissionDenied, "permission denied"),
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
	roles, err := role_service.TestSeedRoles(ctx, 1, test.Core.Role)
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
	roles, err := role_service.TestSeedRoles(ctx, 1, test.Core.Role)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         fmt.Sprintf("/api/v1/roles/%s", roles[0].ID()),
			Method:      http.MethodDelete,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			GotResp:     &errs.Error{},
			ExpResp:     errs.Errorf(errs.PermissionDenied, errs.CodePermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "role-delete-403")
}
