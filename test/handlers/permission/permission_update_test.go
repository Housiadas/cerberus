package permission_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/internal/usecase/permission_usecase"
	"github.com/Housiadas/cerberus/internal/utils/apitest"
	"github.com/Housiadas/cerberus/internal/utils/dbtest"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/pkg/clock"
)

func Test_API_Permission_Update_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_Permission")
	require.NoError(t, err)

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	perms, err := permission_service.TestSeedPermissions(ctx, 1, test.Core.Permission)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         fmt.Sprintf("/api/v1/permissions/%s", perms[0].ID),
			Method:      http.MethodPut,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Input: &permission_usecase.UpdatePermission{
				Name: dbtest.StringPointer("UpdatedPermission"),
			},
			GotResp: &permission_usecase.Permission{},
			ExpResp: &permission_usecase.Permission{
				ID:        perms[0].ID.String(),
				Name:      "UpdatedPermission",
				CreatedAt: clock.Format(&perms[0].CreatedAt),
				UpdatedAt: clock.Format(&perms[0].UpdatedAt),
			},
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

	test, err := apitest.StartTest(t, "Test_API_Permission")
	require.NoError(t, err)

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	perms, err := permission_service.TestSeedPermissions(ctx, 1, test.Core.Permission)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         fmt.Sprintf("/api/v1/permissions/%s", perms[0].ID),
			Method:      http.MethodPut,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &permission_usecase.UpdatePermission{
				Name: dbtest.StringPointer("UpdatedPermission"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Errorf(errs.PermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "permission-update-403")
}
