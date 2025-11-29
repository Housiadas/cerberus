package user_test

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/pkg/errs"
)

func Test_API_User_Create_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_User")
	if err != nil {
		t.Fatalf("Start error: %s", err)
	}

	_, err = insertSeedData(test.DB)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/api/v1/users",
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &user_usecase.NewUser{
				Name:            "Chris Housi",
				Email:           "chris@housi.com",
				Roles:           []string{"ADMIN"},
				Department:      "IT0",
				Password:        "123",
				PasswordConfirm: "123",
			},
			GotResp: &user_usecase.User{},
			ExpResp: &user_usecase.User{
				Name:       "Chris Housi",
				Email:      "chris@housi.com",
				Roles:      []string{"ADMIN"},
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
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	test.Run(t, table, "user-create-200")
}

func Test_API_User_Create_400(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_User")
	if err != nil {
		t.Fatalf("Start error: %s", err)
	}

	_, err = insertSeedData(test.DB)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	table := []apitest.Table{
		{
			Name:       "missing-input",
			URL:        "/api/v1/users",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &user_usecase.NewUser{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "validation: [{\"field\":\"name\",\"error\":\"name is a required field\"},{\"field\":\"email\",\"error\":\"email is a required field\"},{\"field\":\"roles\",\"error\":\"roles is a required field\"},{\"field\":\"password\",\"error\":\"password is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-role",
			URL:        "/api/v1/users",
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &user_usecase.NewUser{
				Name:            "Chris Housi",
				Email:           "chris@housi.com",
				Roles:           []string{"SUPER"},
				Department:      "IT0",
				Password:        "123",
				PasswordConfirm: "123",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "parse: invalid role \"SUPER\""),
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
				Roles:           []string{"USER"},
				Department:      "IT0",
				Password:        "123",
				PasswordConfirm: "123",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "parse: invalid name \"Bi\""),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "user-create-400")
}
