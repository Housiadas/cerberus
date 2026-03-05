package role_test

import (
	"context"
	"net/http"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/internal/usecase/role_usecase"
	"github.com/Housiadas/cerberus/internal/utils/apitest"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/internal/utils/page"
)

func Test_API_Role_Query_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	// Seed extra roles to query
	ctx := context.Background()
	roles, err := role_service.TestSeedRoles(ctx, 2, test.Core.Role)
	require.NoError(t, err)

	sort.Slice(roles, func(i, j int) bool {
		return roles[i].ID.String() <= roles[j].ID.String()
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

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "role-list-forbidden",
			URL:         "/api/v1/roles?page=1&rows=10",
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

	test.Run(t, table, "role-query-403")
}
