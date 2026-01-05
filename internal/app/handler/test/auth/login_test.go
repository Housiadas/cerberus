package auth_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/app/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
)

func Test_API_Auth_Login_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_Auth_Login")
	require.NoError(t, err)

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	usrs := make([]user.User, 0, len(sd.Users))
	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/api/v1/auth/login",
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &auth_usecase.LoginReq{
				Email:    usrs[0].Email.String(),
				Password: "Secret123!@#",
			},
			GotResp: &auth_usecase.Token{},
			ExpResp: &auth_usecase.Token{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*auth_usecase.Token)
				if !exists {
					return "error occurred"
				}

				fmt.Println(gotResp)
				assert.NotEmpty(t, gotResp.AccessToken)
				assert.NotEmpty(t, gotResp.RefreshToken)
				assert.NotEmpty(t, gotResp.ExpiresIn)

				return ""
			},
		},
	}

	test.Run(t, table, "auth-login-200")
}

//func Test_API_Auth_Login_400(t *testing.T) {
//	t.Parallel()
//
//	test, err := apitest.StartTest(t, "Test_API_User")
//	require.NoError(t, err)
//
//	sd, err := insertSeedData(test)
//	require.NoError(t, err)
//
//	usrs := make([]user.User, 0, len(sd.Users))
//	for _, usr := range sd.Users {
//		usrs = append(usrs, usr.User)
//	}
//
//	table := []apitest.Table{
//		{
//			Name:       "missing-password",
//			URL:        "/api/v1/users",
//			Method:     http.MethodPost,
//			StatusCode: http.StatusBadRequest,
//			Input: &auth_usecase.LoginReq{
//				Email: usrs[0].Email.String(),
//			},
//			GotResp: &errs.Error{},
//			ExpResp: errs.Errorf(errs.InvalidArgument, "validate: [{\"field\":\"email\",\"error\":\"mail: no address\"},{\"field\":\"name\",\"error\":\"invalid name \\\"\\\"\"},{\"field\":\"password\",\"error\":\"invalid password \\\"\\\"\"}]"),
//			CmpFunc: func(got any, exp any) string {
//				return cmp.Diff(got, exp)
//			},
//		},
//	}
//
//	test.Run(t, table, "auth-login-400")
//}
