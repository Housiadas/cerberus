package user_test

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

func Test_API_User_Create_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_User")
	require.NoError(t, err)

	_, err = insertSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/api/v1/users",
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &user_usecase.NewUser{
				Name:            "Chris Housi",
				Email:           "chris@housi.com",
				Department:      "IT0",
				Password:        "123",
				PasswordConfirm: "123",
			},
			GotResp: &user_usecase.User{},
			ExpResp: &user_usecase.User{
				Name:       "Chris Housi",
				Email:      "chris@housi.com",
				Department: "IT0",
				Enabled:    true,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*user_usecase.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*user_usecase.User)

				expResp.ID = gotResp.ID
				expResp.CreatedAt = gotResp.CreatedAt
				expResp.UpdatedAt = gotResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	test.Run(t, table, "user-create-200")
}

func Test_API_User_Create_400(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_User")
	require.NoError(t, err)

	_, err = insertSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:       "missing-input",
			URL:        "/api/v1/users",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &user_usecase.NewUser{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Errorf(errs.InvalidArgument, "validate: [{\"field\":\"email\",\"error\":\"mail: no address\"},{\"field\":\"name\",\"error\":\"invalid name \\\"\\\"\"},{\"field\":\"password\",\"error\":\"invalid password \\\"\\\"\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-name",
			URL:        "/api/v1/users",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &user_usecase.NewUser{
				Name:            "Bi",
				Email:           "chris@housi.com",
				Department:      "IT0",
				Password:        "123",
				PasswordConfirm: "123",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Errorf(errs.InvalidArgument, "parse: invalid name \"Bi\""),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "user-create-400")
}
