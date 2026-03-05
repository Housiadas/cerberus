package permission_test

import (
	"context"
	"net/http"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/internal/usecase/permission_usecase"
	"github.com/Housiadas/cerberus/internal/utils/apitest"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/internal/utils/page"
)

func Test_API_Permission_Query_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_Permission")
	require.NoError(t, err)

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	ctx := context.Background()
	perms, err := permission_service.TestSeedPermissions(ctx, 2, test.Core.Permission)
	require.NoError(t, err)

	sort.Slice(perms, func(i, j int) bool {
		return perms[i].ID.String() <= perms[j].ID.String()
	})

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         "/api/v1/permissions?page=1&rows=10&orderBy=id,ASC&name=Permission",
			Method:      http.MethodGet,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			GotResp:     &page.Result[permission_usecase.Permission]{},
			ExpResp: &page.Result[permission_usecase.Permission]{
				Data: toAppPermissions(perms),
				Metadata: page.Metadata{
					FirstPage:   1,
					CurrentPage: 1,
					LastPage:    1,
					RowsPerPage: 10,
					Total:       len(perms),
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

	test, err := apitest.StartTest(t, "Test_API_Permission")
	require.NoError(t, err)

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "permission-list-forbidden",
			URL:         "/api/v1/permissions?page=1&rows=10",
			Method:      http.MethodGet,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			GotResp:     &errs.Error{},
			ExpResp:     errs.Errorf(errs.PermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "permission-query-403")
}
