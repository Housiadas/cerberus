package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/Housiadas/cerberus/internal/core/permission"
	errs "github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/apitest"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/dbtest"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func Test_API_Permission_Query_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertPermissionSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	perms, err := permission.TestSeedPermissions(ctx, 2, test.Core.Permission)
	require.NoError(t, err)

	sort.Slice(perms, func(i, j int) bool {
		return perms[i].ID().String() <= perms[j].ID().String()
	})

	expMd := openapi.Metadata{
		HasMore: false,
		Limit:   10,
	}

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         "/api/v1/permissions?limit=10&orderBy=id,ASC&name=Permission",
			Method:      http.MethodGet,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			GotResp:     &openapi.PermissionPageResult{},
			ExpResp: &openapi.PermissionPageResult{
				Data:     new(toTestPermissions(perms)),
				Metadata: &expMd,
			},
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "permission-query-200")
}

func Test_API_Permission_Query_403(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertPermissionSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "permission-list-forbidden",
			URL:         "/api/v1/permissions?limit=10",
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

	test.Run(t, table, "permission-query-403")
}

func Test_API_Permission_Create_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertPermissionSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         "/api/v1/permissions",
			Method:      http.MethodPost,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Input: &openapi.NewPermission{
				Name: "document:read",
			},
			GotResp: &openapi.Permission{},
			ExpResp: &openapi.Permission{
				Name: "document:read",
			},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*openapi.Permission)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*openapi.Permission)
				expResp.Id = gotResp.Id
				expResp.CreatedAt = gotResp.CreatedAt
				expResp.UpdatedAt = gotResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	test.Run(t, table, "permission-create-200")
}

func Test_API_Permission_Create_400(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertPermissionSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "missing-name",
			URL:         "/api/v1/permissions",
			Method:      http.MethodPost,
			StatusCode:  http.StatusBadRequest,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Input:       &openapi.NewPermission{},
			GotResp:     &errs.Error{},
			ExpResp: &errs.Error{
				Status:  errs.InvalidArgument,
				Code:    errs.CodeValidation,
				Message: `[{"field":"name","error":"name is required"}]`,
				Fields:  []errs.FieldError{{Field: "name", Err: "name is required"}},
			},
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-access-token",
			URL:        "/api/v1/permissions",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input:      &openapi.NewPermission{Name: "document:read"},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Errorf(errs.Unauthenticated, errs.CodeUnauthenticated, "expected authorization header format: Bearer <token>"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "permission-create-400")
}

func Test_API_Permission_Create_403(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertPermissionSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         "/api/v1/permissions",
			Method:      http.MethodPost,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input:       &openapi.NewPermission{Name: "document:read"},
			GotResp:     &errs.Error{},
			ExpResp:     errs.Errorf(errs.PermissionDenied, errs.CodePermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "permission-create-403")
}

func Test_API_Permission_Update_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertPermissionSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	perms, err := permission.TestSeedPermissions(ctx, 1, test.Core.Permission)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         fmt.Sprintf("/api/v1/permissions/%s", perms[0].ID()),
			Method:      http.MethodPut,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Input: &openapi.UpdatePermission{
				Name: dbtest.StringPointer("UpdatedPermission"),
			},
			GotResp: &openapi.Permission{},
			ExpResp: func() *openapi.Permission {
				return &openapi.Permission{
					Id:        perms[0].ID().String(),
					Name:      "UpdatedPermission",
					CreatedAt: clock.Format(new(perms[0].CreatedAt())),
					UpdatedAt: clock.Format(new(perms[0].UpdatedAt())),
				}
			}(),
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*openapi.Permission)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*openapi.Permission)
				gotResp.UpdatedAt = expResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	test.Run(t, table, "permission-update-200")
}

func Test_API_Permission_Update_403(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertPermissionSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	perms, err := permission.TestSeedPermissions(ctx, 1, test.Core.Permission)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         fmt.Sprintf("/api/v1/permissions/%s", perms[0].ID()),
			Method:      http.MethodPut,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &openapi.UpdatePermission{
				Name: dbtest.StringPointer("UpdatedPermission"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Errorf(errs.PermissionDenied, errs.CodePermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "permission-update-403")
}

func Test_API_Permission_Delete_204(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertPermissionSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	perms, err := permission.TestSeedPermissions(ctx, 1, test.Core.Permission)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         fmt.Sprintf("/api/v1/permissions/%s", perms[0].ID()),
			Method:      http.MethodDelete,
			StatusCode:  http.StatusNoContent,
			AccessToken: &sd.Admins[0].AccessToken.Token,
		},
	}

	test.Run(t, table, "permission-delete-204")
}

func Test_API_Permission_Delete_403(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertPermissionSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	perms, err := permission.TestSeedPermissions(ctx, 1, test.Core.Permission)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         fmt.Sprintf("/api/v1/permissions/%s", perms[0].ID()),
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

	test.Run(t, table, "permission-delete-403")
}
