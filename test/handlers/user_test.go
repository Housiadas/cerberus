package handlers_test

import (
	"fmt"
	"net/http"
	"sort"
	"testing"

	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/testutil/apitest"
	"github.com/Housiadas/cerberus/internal/testutil/dbtest"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/cursor"
)

func Test_API_User_GetMe_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertUserSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "as-user",
			URL:         "/api/v1/users/me",
			StatusCode:  http.StatusOK,
			Method:      http.MethodGet,
			AccessToken: &sd.Users[0].AccessToken.Token,
			GotResp:     &user_usecase.User{},
			ExpResp:     toAppUserPtr(sd.Users[0].User),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:        "as-admin",
			URL:         "/api/v1/users/me",
			StatusCode:  http.StatusOK,
			Method:      http.MethodGet,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			GotResp:     &user_usecase.User{},
			ExpResp:     toAppUserPtr(sd.Admins[0].User),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "get-me-200")
}

func Test_API_User_GetMe_401(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:       "missing-access-token",
			URL:        "/api/v1/users/me",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &errs2.Error{},
			ExpResp:    errs2.Errorf(errs2.Unauthenticated, errs2.CodeUnauthenticated, "expected authorization header format: Bearer <token>"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "get-me-401")
}

func Test_API_User_UpdateMe_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertUserSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         "/api/v1/users/me",
			Method:      http.MethodPut,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &user_usecase.UpdateMe{
				Name:       dbtest.StringPointer("Updated Name"),
				Email:      dbtest.StringPointer("updated@housi.com"),
				Department: dbtest.StringPointer("Engineering"),
			},
			GotResp: &user_usecase.User{},
			ExpResp: &user_usecase.User{
				ID:         sd.Users[0].ID().String(),
				Name:       "Updated Name",
				Email:      "updated@housi.com",
				Department: "Engineering",
				Enabled:    true,
				CreatedAt:  func() string { ct := sd.Users[0].CreatedAt(); return clock.Format(&ct) }(),
			},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*user_usecase.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*user_usecase.User)
				expResp.UpdatedAt = gotResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	test.Run(t, table, "update-me-200")
}

func Test_API_User_UpdateMe_400(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertUserSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "bad-email",
			URL:         "/api/v1/users/me",
			Method:      http.MethodPut,
			StatusCode:  http.StatusBadRequest,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &user_usecase.UpdateMe{
				Email:           dbtest.StringPointer("bill@"),
				PasswordConfirm: dbtest.StringPointer("jack"),
			},
			GotResp: &errs2.Error{},
			ExpResp: errs2.Errorf(errs2.InvalidArgument, errs2.CodeValidation, "validate: [{\"field\":\"email\",\"error\":\"mail: missing '@' or angle-addr\"},{\"field\":\"password\",\"error\":\"passwords do not match\"}]"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "update-me-400")
}

func Test_API_User_UpdateMe_401(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:       "missing-access-token",
			URL:        "/api/v1/users/me",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPut,
			Input: &user_usecase.UpdateMe{
				Name: dbtest.StringPointer("Test"),
			},
			GotResp: &errs2.Error{},
			ExpResp: errs2.Errorf(errs2.Unauthenticated, errs2.CodeUnauthenticated, "expected authorization header format: Bearer <token>"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "update-me-401")
}

func Test_API_User_Query_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertUserSeedData(test)
	require.NoError(t, err)

	usrs := make([]user.User, 0, len(sd.Admins)+len(sd.Users))
	for _, usr := range sd.Admins {
		usrs = append(usrs, usr.User)
	}
	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	sort.Slice(usrs, func(i, j int) bool {
		return usrs[i].ID().String() <= usrs[j].ID().String()
	})

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         "/api/v1/users?limit=10&orderBy=user_id,ASC&name=Name",
			StatusCode:  http.StatusOK,
			Method:      http.MethodGet,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			GotResp:     &cursor.Result[user_usecase.User]{},
			ExpResp: &cursor.Result[user_usecase.User]{
				Data: toAppUsers(usrs),
				Metadata: cursor.Metadata{
					HasMore: false,
					Limit:   10,
				},
			},
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "user-query-200")
}

func Test_API_User_Query_BY_ID_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertUserSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         fmt.Sprintf("/api/v1/users/%s", sd.Users[0].ID()),
			StatusCode:  http.StatusOK,
			Method:      http.MethodGet,
			AccessToken: &sd.Users[0].AccessToken.Token,
			GotResp:     &user_usecase.User{},
			ExpResp:     toAppUserPtr(sd.Users[0].User),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "user-query-by-id-200")
}

func Test_API_User_Query_400(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertUserSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "bad-query-filter",
			URL:         "/api/v1/users?limit=10&email=a.com",
			StatusCode:  http.StatusBadRequest,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Method:      http.MethodGet,
			GotResp:     &errs2.Error{},
			ExpResp:     errs2.Errorf(errs2.InvalidArgument, errs2.CodeValidation, "[{\"field\":\"email\",\"error\":\"mail: missing '@' or angle-addr\"}]"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:        "bad-order-by-value",
			URL:         "/api/v1/users?limit=10&orderBy=ser_id,ASC",
			StatusCode:  http.StatusBadRequest,
			Method:      http.MethodGet,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			GotResp:     &errs2.Error{},
			ExpResp:     errs2.Errorf(errs2.InvalidArgument, errs2.CodeValidation, "[{\"field\":\"order\",\"error\":\"unknown order: ser_id\"}]"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "user-query-400")
}

func Test_API_User_Query_403(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertUserSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "user-list-forbidden",
			URL:         "/api/v1/users?limit=10",
			StatusCode:  http.StatusForbidden,
			Method:      http.MethodGet,
			AccessToken: &sd.Users[0].AccessToken.Token,
			GotResp:     &errs2.Error{},
			ExpResp:     errs2.Errorf(errs2.PermissionDenied, errs2.CodePermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "user-query-403")
}

func Test_API_User_Create_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertUserSeedData(test)
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
			AccessToken: &sd.Admins[0].AccessToken.Token,
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

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertUserSeedData(test)
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
			GotResp:    &errs2.Error{},
			ExpResp:    errs2.Errorf(errs2.Unauthenticated, errs2.CodeUnauthenticated, "expected authorization header format: Bearer <token>"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:        "missing-input",
			URL:         "/api/v1/users",
			Method:      http.MethodPost,
			StatusCode:  http.StatusBadRequest,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Input:       &user_usecase.NewUser{},
			GotResp:     &errs2.Error{},
			ExpResp:     &errs2.Error{},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs2.Error)
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
			AccessToken: &sd.Admins[0].AccessToken.Token,
			Input: &user_usecase.NewUser{
				Name:            "Bi",
				Email:           "chris@housi.com",
				Department:      "IT0",
				Password:        "123",
				PasswordConfirm: "123",
			},
			GotResp: &errs2.Error{},
			ExpResp: errs2.Errorf(errs2.InvalidArgument, errs2.CodeValidation, "validate: [{\"field\":\"name\",\"error\":\"invalid name value: \\\"Bi\\\"\"}]"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "user-create-400")
}

func Test_API_User_Update_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertUserSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         fmt.Sprintf("/api/v1/users/%s", sd.Users[0].ID()),
			Method:      http.MethodPut,
			StatusCode:  http.StatusOK,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &user_usecase.UpdateUser{
				Name:            dbtest.StringPointer("Jack Housi"),
				Email:           dbtest.StringPointer("chris@housi2.com"),
				Department:      dbtest.StringPointer("IT0"),
				Password:        dbtest.StringPointer("123"),
				PasswordConfirm: dbtest.StringPointer("123"),
			},
			GotResp: &user_usecase.User{},
			ExpResp: &user_usecase.User{
				ID:         sd.Users[0].ID().String(),
				Name:       "Jack Housi",
				Email:      "chris@housi2.com",
				Department: "IT0",
				Enabled:    true,
				CreatedAt:  func() string { return clock.Format(new(sd.Users[0].CreatedAt())) }(),
				UpdatedAt:  func() string { return clock.Format(new(sd.Users[0].UpdatedAt())) }(),
			},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*user_usecase.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*user_usecase.User)
				gotResp.UpdatedAt = expResp.UpdatedAt

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	test.Run(t, table, "update-200")
}

func Test_API_User_Update_400(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	if err != nil {
		t.Fatalf("Start error: %s", err)
	}

	sd, err := insertUserSeedData(test)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	table := []apitest.Table{
		{
			Name:        "bad-input",
			URL:         fmt.Sprintf("/api/v1/users/%s", sd.Users[0].ID()),
			Method:      http.MethodPut,
			StatusCode:  http.StatusBadRequest,
			AccessToken: &sd.Users[0].AccessToken.Token,
			Input: &user_usecase.UpdateUser{
				Email:           dbtest.StringPointer("bill@"),
				PasswordConfirm: dbtest.StringPointer("jack"),
			},
			GotResp: &errs2.Error{},
			ExpResp: errs2.Errorf(errs2.InvalidArgument, errs2.CodeValidation, "validate: [{\"field\":\"email\",\"error\":\"mail: missing '@' or angle-addr\"},{\"field\":\"password\",\"error\":\"passwords do not match\"}]"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "update-400")
}

func Test_API_User_Delete_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertUserSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "asadmin",
			URL:         fmt.Sprintf("/api/v1/users/%s", sd.Users[0].ID()),
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

	sd, err := insertUserSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "user-delete-forbidden",
			URL:         fmt.Sprintf("/api/v1/users/%s", sd.Admins[0].ID()),
			Method:      http.MethodDelete,
			StatusCode:  http.StatusForbidden,
			AccessToken: &sd.Users[0].AccessToken.Token,
			GotResp:     &errs2.Error{},
			ExpResp:     errs2.Errorf(errs2.PermissionDenied, errs2.CodePermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "delete-403")
}
