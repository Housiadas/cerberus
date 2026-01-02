package user_test

import (
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

func Test_API_User_Query_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_User")
	require.NoError(t, err)

	sd, err := insertSeedData(test.DB)
	require.NoError(t, err)

	usrs := make([]user.User, 0, len(sd.Admins)+len(sd.Users))
	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	sort.Slice(usrs, func(i, j int) bool {
		return usrs[i].ID.String() <= usrs[j].ID.String()
	})

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/api/v1/users?page=1&rows=10&orderBy=user_id,ASC&name=Name",
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &web.Result[user_usecase.User]{},
			ExpResp: &web.Result[user_usecase.User]{
				Data: toAppUsers(usrs),
				Metadata: web.Metadata{
					FirstPage:   1,
					CurrentPage: 1,
					LastPage:    1,
					RowsPerPage: 10,
					Total:       len(usrs),
				},
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "user-query-200")
}

func Test_API_User_Query_BY_ID_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_User")
	require.NoError(t, err)

	sd, err := insertSeedData(test.DB)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/api/v1/users/%s", sd.Users[0].ID),
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &user_usecase.User{},
			ExpResp:    toAppUserPtr(sd.Users[0].User),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "user-query-by-id-200")
}

func Test_API_User_Query_400(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_User")
	require.NoError(t, err)

	_, err = insertSeedData(test.DB)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:       "bad-query-filter",
			URL:        "/api/v1/users?page=1&rows=10&email=a.com",
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Errorf(errs.InvalidArgument, "[{\"field\":\"email\",\"error\":\"mail: missing '@' or angle-addr\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-order-by-value",
			URL:        "/api/v1/users?page=1&rows=10&orderBy=ser_id,ASC",
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Errorf(errs.InvalidArgument, "[{\"field\":\"order\",\"error\":\"unknown order: ser_id\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "user-query-400")
}
