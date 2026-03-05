package role_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/internal/usecase/role_usecase"
	"github.com/Housiadas/cerberus/internal/utils/apitest"
	"github.com/Housiadas/cerberus/internal/utils/dbtest"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/pkg/clock"
)

func Test_API_Role_Update_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_Role")
	require.NoError(t, err)

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	roles, err := role_service.TestSeedRoles(ctx, 1, test.Core.Role)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         fmt.Sprintf("/api/v1/roles/%s", roles[0].ID),
			Method:      http.MethodPut,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Input: &role_usecase.UpdateRole{
				Name: dbtest.StringPointer("UpdatedRole"),
			},
			GotResp: &role_usecase.Role{},
			ExpResp: &role_usecase.Role{
				ID:        roles[0].ID.String(),
				Name:      "UpdatedRole",
				CreatedAt: clock.Format(&roles[0].CreatedAt),
				UpdatedAt: clock.Format(&roles[0].UpdatedAt),
			},
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

	test, err := apitest.StartTest(t, "Test_API_Role")
	require.NoError(t, err)

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	roles, err := role_service.TestSeedRoles(ctx, 1, test.Core.Role)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         fmt.Sprintf("/api/v1/roles/%s", roles[0].ID),
			Method:      http.MethodPut,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &role_usecase.UpdateRole{
				Name: dbtest.StringPointer("UpdatedRole"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Errorf(errs.PermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "role-update-403")
}
