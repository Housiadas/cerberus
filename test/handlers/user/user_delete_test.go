package user_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/utils/apitest"
	"github.com/Housiadas/cerberus/internal/utils/errs"
)

func Test_API_User_Delete_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "asadmin",
			URL:         fmt.Sprintf("/api/v1/users/%s", sd.Users[0].ID),
			Method:      http.MethodDelete,
			StatusCode:  http.StatusNoContent,
			AccessToken: &sd.Admins[0].AccessToken.Token,
		},
	}

	test.Run(t, table, "delete-200")
}

func Test_API_User_Delete_403(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "user-delete-forbidden",
			URL:         fmt.Sprintf("/api/v1/users/%s", sd.Admins[0].ID),
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

	test.Run(t, table, "delete-403")
}
