package user_test

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

func Test_API_User_Create_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_User")
	require.NoError(t, err)

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	usrs := make([]user.User, 0, len(sd.Users))
	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         "/api/v1/users",
			Method:      http.MethodPost,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Users[0].AccessToken.Token,
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
			AssertFunc: func(got any, exp any) string {
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

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	usrs := make([]user.User, 0, len(sd.Users))
	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	table := []apitest.Table{
		{
			Name:       "missing-access-token",
			URL:        "/api/v1/users",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input:      &user_usecase.NewUser{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Errorf(errs.Unauthenticated, "expected authorization header format: Bearer <token>"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:        "missing-input",
			URL:         "/api/v1/users",
			Method:      http.MethodPost,
			StatusCode:  http.StatusBadRequest,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input:       &user_usecase.NewUser{},
			GotResp:     &errs.Error{},
			ExpResp:     &errs.Error{},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}

				assert.Len(t, gotResp.Fields, 4)
				assert.Contains(t, gotResp.Fields[0].Field, "name")
				return ""
			},
		},
		{
			Name:        "bad-name",
			URL:         "/api/v1/users",
			Method:      http.MethodPost,
			StatusCode:  http.StatusBadRequest,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &user_usecase.NewUser{
				Name:            "Bi",
				Email:           "chris@housi.com",
				Department:      "IT0",
				Password:        "123",
				PasswordConfirm: "123",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Errorf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"invalid name \\\"Bi\\\"\"}]"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "user-create-400")
}
