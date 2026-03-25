package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"testing"

	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/testutil/apitest"
	"github.com/Housiadas/cerberus/internal/testutil/dbtest"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/internal/usecase/permission_usecase"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/cursor"
)

func Test_API_Permission_Query_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertPermissionSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	perms, err := permission_service.TestSeedPermissions(ctx, 2, test.Core.Permission)
	require.NoError(t, err)

	sort.Slice(perms, func(i, j int) bool {
		return perms[i].ID().String() <= perms[j].ID().String()
	})

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         "/api/v1/permissions?limit=10&orderBy=id,ASC&name=Permission",
			Method:      http.MethodGet,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			GotResp:     &cursor.Result[permission_usecase.Permission]{},
			ExpResp: &cursor.Result[permission_usecase.Permission]{
				Data: toAppPermissions(perms),
				Metadata: cursor.Metadata{
					HasMore: false,
					Limit:   10,
				},
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
			GotResp:     &errs2.Error{},
			ExpResp:     errs2.Errorf(errs2.PermissionDenied, errs2.CodePermissionDenied, "permission denied"),
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
			Input: &permission_usecase.NewPermission{
				Name: "document:read",
			},
			GotResp: &permission_usecase.Permission{},
			ExpResp: &permission_usecase.Permission{
				Name: "document:read",
			},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*permission_usecase.Permission)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*permission_usecase.Permission)
				expResp.ID = gotResp.ID
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
			Input:       &permission_usecase.NewPermission{},
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
			URL:        "/api/v1/permissions",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input:      &permission_usecase.NewPermission{Name: "document:read"},
			GotResp:    &errs2.Error{},
			ExpResp:    errs2.Errorf(errs2.Unauthenticated, errs2.CodeUnauthenticated, "expected authorization header format: Bearer <token>"),
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
			Input:       &permission_usecase.NewPermission{Name: "document:read"},
			GotResp:     &errs2.Error{},
			ExpResp:     errs2.Errorf(errs2.PermissionDenied, errs2.CodePermissionDenied, "permission denied"),
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
	perms, err := permission_service.TestSeedPermissions(ctx, 1, test.Core.Permission)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         fmt.Sprintf("/api/v1/permissions/%s", perms[0].ID()),
			Method:      http.MethodPut,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Input: &permission_usecase.UpdatePermission{
				Name: dbtest.StringPointer("UpdatedPermission"),
			},
			GotResp: &permission_usecase.Permission{},
			ExpResp: func() *permission_usecase.Permission {
				return &permission_usecase.Permission{
					ID:        perms[0].ID().String(),
					Name:      "UpdatedPermission",
					CreatedAt: clock.Format(new(perms[0].CreatedAt())),
					UpdatedAt: clock.Format(new(perms[0].UpdatedAt())),
				}
			}(),
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*permission_usecase.Permission)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*permission_usecase.Permission)
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
	perms, err := permission_service.TestSeedPermissions(ctx, 1, test.Core.Permission)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         fmt.Sprintf("/api/v1/permissions/%s", perms[0].ID()),
			Method:      http.MethodPut,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &permission_usecase.UpdatePermission{
				Name: dbtest.StringPointer("UpdatedPermission"),
			},
			GotResp: &errs2.Error{},
			ExpResp: errs2.Errorf(errs2.PermissionDenied, errs2.CodePermissionDenied, "permission denied"),
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
	perms, err := permission_service.TestSeedPermissions(ctx, 1, test.Core.Permission)
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
	perms, err := permission_service.TestSeedPermissions(ctx, 1, test.Core.Permission)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         fmt.Sprintf("/api/v1/permissions/%s", perms[0].ID()),
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

	test.Run(t, table, "permission-delete-403")
}
