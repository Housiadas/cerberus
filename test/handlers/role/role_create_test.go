package role_test

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/usecase/role_usecase"
	"github.com/Housiadas/cerberus/internal/utils/apitest"
	"github.com/Housiadas/cerberus/internal/utils/errs"
)

func Test_API_Role_Create_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertSeedData(test)
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

	sd, err := insertSeedData(test)
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
			ExpResp:     errs.Errorf(errs.InvalidArgument, "parse: invalid name value: \"\""),
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
			ExpResp:    errs.Errorf(errs.Unauthenticated, "expected authorization header format: Bearer <token>"),
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

	sd, err := insertSeedData(test)
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
			ExpResp:     errs.Errorf(errs.PermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "role-create-403")
}
