package handlers_test

import (
	"net/http"
	"sort"
	"strings"
	"testing"

	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/testutil/apitest"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/pkg/cursor"
)

func Test_API_Audit_Query_200(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertAuditSeedData(test)
	require.NoError(t, err)

	sort.Slice(sd.Admins[0].Audits, func(i, j int) bool {
		return sd.Admins[0].Audits[i].ObjName().String() <= sd.Admins[0].Audits[j].ObjName().String()
	})

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         "/api/v1/audits?limit=10&orderBy=obj_name,ASC&obj_name=ObjName",
			StatusCode:  http.StatusOK,
			Method:      http.MethodGet,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			GotResp:     &cursor.Result[audit_usecase.Audit]{},
			ExpResp: &cursor.Result[audit_usecase.Audit]{
				Metadata: cursor.Metadata{
					HasMore: false,
					Limit:   10,
				},
				Data: toAppAudits(sd.Admins[0].Audits),
			},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*cursor.Result[audit_usecase.Audit])
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*cursor.Result[audit_usecase.Audit])

				for i := range gotResp.Data {
					if gotResp.Data[i].Timestamp == expResp.Data[i].Timestamp {
						expResp.Data[i].Timestamp = gotResp.Data[i].Timestamp
					}

					gotResp.Data[i].Data = strings.ReplaceAll(gotResp.Data[i].Data, " ", "")
					expResp.Data[i].Data = strings.ReplaceAll(expResp.Data[i].Data, " ", "")
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	test.Run(t, table, "audit-query-200")
}

func Test_API_Audit_Query_400(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertAuditSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "bad-query-filter",
			URL:         "/api/v1/audits?limit=10&obj_id=123",
			StatusCode:  http.StatusBadRequest,
			Method:      http.MethodGet,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			GotResp:     &errs2.Error{},
			ExpResp:     errs2.Errorf(errs2.InvalidArgument, errs2.CodeValidation, "[{\"field\":\"obj_id\",\"error\":\"invalid UUID length: 3\"}]"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:        "bad-order-by-value",
			URL:         "/api/v1/audits?limit=10&orderBy=ser_id,ASC",
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

	test.Run(t, table, "audit-query-400")
}

func Test_API_Audit_Query_403(t *testing.T) {
	t.Parallel()

	test, err := env.StartTest(t, t.Name())
	require.NoError(t, err)

	sd, err := insertAuditSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:        "user-audit-forbidden",
			URL:         "/api/v1/audits?limit=10",
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

	test.Run(t, table, "audit-query-403")
}
