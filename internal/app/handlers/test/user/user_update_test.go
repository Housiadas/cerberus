package user_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/internal/common/dbtest"
	"github.com/Housiadas/cerberus/pkg/errs"
)

func Test_API_User_Update_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_User")
	if err != nil {
		t.Fatalf("Start error: %s", err)
	}

	sd, err := insertSeedData(test.DB)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/api/v1/users/%s", sd.Users[0].ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &user_usecase.UpdateUser{
				Name:            dbtest.StringPointer("Jack Housi"),
				Email:           dbtest.StringPointer("chris@housi2.com"),
				Department:      dbtest.StringPointer("IT0"),
				Password:        dbtest.StringPointer("123"),
				PasswordConfirm: dbtest.StringPointer("123"),
			},
			GotResp: &user_usecase.User{},
			ExpResp: &user_usecase.User{
				ID:          sd.Users[0].ID.String(),
				Name:        "Jack Housi",
				Email:       "chris@housi2.com",
				Roles:       []string{"USER"},
				Department:  "IT0",
				Enabled:     true,
				DateCreated: sd.Users[0].DateCreated.Format(time.RFC3339),
				DateUpdated: sd.Users[0].DateUpdated.Format(time.RFC3339),
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*user_usecase.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*user_usecase.User)
				gotResp.DateUpdated = expResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "role",
			URL:        fmt.Sprintf("/api/v1/users/role/%s", sd.Admins[0].ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &user_usecase.UpdateUserRole{
				Roles: []string{"USER"},
			},
			GotResp: &user_usecase.User{},
			ExpResp: &user_usecase.User{
				ID:          sd.Admins[0].ID.String(),
				Name:        sd.Admins[0].Name.String(),
				Email:       sd.Admins[0].Email.Address,
				Roles:       []string{"USER"},
				Department:  sd.Admins[0].Department.String(),
				Enabled:     true,
				DateCreated: sd.Admins[0].DateCreated.Format(time.RFC3339),
				DateUpdated: sd.Admins[0].DateUpdated.Format(time.RFC3339),
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*user_usecase.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*user_usecase.User)
				gotResp.DateUpdated = expResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	test.Run(t, table, "update-200")
}

func Test_API_User_Update_400(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_User")
	if err != nil {
		t.Fatalf("Start error: %s", err)
	}

	sd, err := insertSeedData(test.DB)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	table := []apitest.Table{
		{
			Name:       "bad-input",
			URL:        fmt.Sprintf("/api/v1/users/%s", sd.Users[0].ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &user_usecase.UpdateUser{
				Email:           dbtest.StringPointer("bill@"),
				PasswordConfirm: dbtest.StringPointer("jack"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validation: [{\"field\":\"email\",\"error\":\"email must be a valid email address\"},{\"field\":\"passwordConfirm\",\"error\":\"passwordConfirm must be equal to Password\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-role",
			URL:        fmt.Sprintf("/api/v1/users/role/%s", sd.Admins[0].ID),
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &user_usecase.UpdateUserRole{
				Roles: []string{"BAD ROLE"},
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "parse: invalid role \"BAD ROLE\""),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "update-400")
}
