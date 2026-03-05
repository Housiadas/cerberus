package permission_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/internal/utils/apitest"
	"github.com/Housiadas/cerberus/internal/utils/errs"
)

func Test_API_Permission_Delete_204(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
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

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	perms, err := permission_service.TestSeedPermissions(ctx, 1, test.Core.Permission)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "forbidden",
			URL:         fmt.Sprintf("/api/v1/permissions/%s", perms[0].ID),
			Method:      http.MethodDelete,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			GotResp:     &errs.Error{},
			ExpResp:     errs.Errorf(errs.PermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "permission-delete-403")
}
