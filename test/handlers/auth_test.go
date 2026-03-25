package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/testutil/apitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/usecase/auth_usecase"
)

func Test_API_Auth_Login_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertAuthSeedData(test)
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
				Email:    usrs[0].Email().Address,
				Password: "Secret123!@#",
			},
			GotResp: &auth_usecase.Token{},
			ExpResp: &auth_usecase.Token{},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*auth_usecase.Token)
				if !exists {
					return "error occurred"
				}

				assert.NotEmpty(t, gotResp.AccessToken)
				assert.NotEmpty(t, gotResp.RefreshToken)
				assert.NotEmpty(t, gotResp.ExpiresIn)

				return ""
			},
		},
	}

	test.Run(t, table, "auth-login-200")
}

func Test_API_Auth_Login_400(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertAuthSeedData(test)
	require.NoError(t, err)

	usrs := make([]user.User, 0, len(sd.Users))
	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	table := []apitest.Table{
		{
			Name:       "missing-password",
			URL:        "/api/v1/auth/login",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &auth_usecase.LoginReq{
				Email: usrs[0].Email().Address,
			},
			GotResp: &errs.Error{},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}

				assert.Len(t, gotResp.Fields, 1)
				assert.Contains(t, gotResp.Fields[0].Field, "password")
				return ""
			},
		},
	}

	test.Run(t, table, "auth-login-400")
}

func Test_API_Auth_ForgotPassword_204(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertAuthSeedData(test)
	require.NoError(t, err)

	usrs := make([]user.User, 0, len(sd.Users))
	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	table := []apitest.Table{
		{
			Name:       "existing-email",
			URL:        "/api/v1/auth/forgot-password",
			Method:     http.MethodPost,
			StatusCode: http.StatusNoContent,
			Input: &auth_usecase.ForgotPasswordReq{
				Email: usrs[0].Email().Address,
			},
		},
		{
			Name:       "nonexistent-email-still-204",
			URL:        "/api/v1/auth/forgot-password",
			Method:     http.MethodPost,
			StatusCode: http.StatusNoContent,
			Input: &auth_usecase.ForgotPasswordReq{
				Email: "nonexistent@example.com",
			},
		},
	}

	test.Run(t, table, "auth-forgot-password-204")
}

func Test_API_Auth_ForgotPassword_400(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:       "missing-email",
			URL:        "/api/v1/auth/forgot-password",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &auth_usecase.ForgotPasswordReq{},
			GotResp:    &errs.Error{},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}

				assert.NotEmpty(t, gotResp.Fields)
				return ""
			},
		},
	}

	test.Run(t, table, "auth-forgot-password-400")
}

func Test_API_Auth_ResetPassword_204(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertAuthSeedData(test)
	require.NoError(t, err)

	usrs := make([]user.User, 0, len(sd.Users))
	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	// Create a reset token for the user
	tkn, err := test.Core.ResetToken.Create(ctx, usrs[0].ID())
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "valid-reset",
			URL:         "/api/v1/auth/reset-password",
			Method:      http.MethodPost,
			StatusCode:  http.StatusNoContent,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &auth_usecase.ResetPasswordReq{
				Token:           tkn.Token(),
				OldPassword:     "Secret123!@#",
				Password:        "NewSecret456!@#",
				PasswordConfirm: "NewSecret456!@#",
			},
		},
	}

	test.Run(t, table, "auth-reset-password-204")
}

func Test_API_Auth_ResetPassword_400(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertAuthSeedData(test)
	require.NoError(t, err)

	usrs := make([]user.User, 0, len(sd.Users))
	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	// Create a reset token to use for wrong-password test
	tkn, err := test.Core.ResetToken.Create(ctx, usrs[0].ID())
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "invalid-token",
			URL:         "/api/v1/auth/reset-password",
			Method:      http.MethodPost,
			StatusCode:  http.StatusNotFound,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &auth_usecase.ResetPasswordReq{
				Token:           "invalid-token",
				OldPassword:     "Secret123!@#",
				Password:        "NewSecret456!@#",
				PasswordConfirm: "NewSecret456!@#",
			},
			GotResp: &errs.Error{},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				assert.NotEmpty(t, gotResp.Message)
				return ""
			},
		},
		{
			Name:        "wrong-old-password",
			URL:         "/api/v1/auth/reset-password",
			Method:      http.MethodPost,
			StatusCode:  http.StatusBadRequest,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &auth_usecase.ResetPasswordReq{
				Token:           tkn.Token(),
				OldPassword:     "WrongPassword123!",
				Password:        "NewSecret456!@#",
				PasswordConfirm: "NewSecret456!@#",
			},
			GotResp: &errs.Error{},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				assert.NotEmpty(t, gotResp.Message)
				return ""
			},
		},
	}

	test.Run(t, table, "auth-reset-password-400")
}
