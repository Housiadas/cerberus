package user_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/pkg/web/errs"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/internal/common/dbtest"
)

func Test_API_User_Update_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_User")
	require.NoError(t, err)

	sd, err := insertSeedData(test.DB)
	require.NoError(t, err)

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
				ID:         sd.Users[0].ID.String(),
				Name:       "Jack Housi",
				Email:      "chris@housi2.com",
				Department: "IT0",
				Enabled:    true,
				CreatedAt:  sd.Users[0].CreatedAt.Format(time.RFC3339),
				UpdatedAt:  sd.Users[0].UpdatedAt.Format(time.RFC3339),
			},
			CmpFunc: func(got any, exp any) string {
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
			ExpResp: errs.Errorf(errs.InvalidArgument, "validation: [{\"field\":\"email\",\"error\":\"email must be a valid email address\"},{\"field\":\"passwordConfirm\",\"error\":\"passwordConfirm must be equal to Password\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "update-400")
}
