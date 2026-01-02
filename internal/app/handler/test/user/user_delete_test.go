package user_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/common/apitest"
)

func Test_API_User_Delete_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_User")
	require.NoError(t, err)

	sd, err := insertSeedData(test.DB)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:       "asuser",
			URL:        fmt.Sprintf("/api/v1/users/%s", sd.Users[1].ID),
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
		},
	}

	test.Run(t, table, "delete-200")
}
