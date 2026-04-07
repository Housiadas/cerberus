package handlers_test

import (
	"net/http"
	"sort"
	"testing"

	"github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/apitest"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
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

	expMd := openapi.Metadata{
		HasMore: false,
		Limit:   10,
	}

	table := []apitest.Table{
		{
			Name:        "basic",
			URL:         "/api/v1/audits?limit=10&orderBy=obj_name,ASC&obj_name=ObjName",
			StatusCode:  http.StatusOK,
			Method:      http.MethodGet,
			AccessToken: &sd.Admins[0].AccessToken.Token,
			GotResp:     &openapi.AuditPageResult{},
			ExpResp: &openapi.AuditPageResult{
				Metadata: &expMd,
				Data:     new(toTestAudits(sd.Admins[0].Audits)),
			},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*openapi.AuditPageResult)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*openapi.AuditPageResult)
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
			GotResp:     &errs.Error{},
			ExpResp:     errs.Errorf(errs.InvalidArgument, errs.CodeValidation, "[{\"field\":\"obj_id\",\"error\":\"invalid UUID length: 3\"}]"),
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
			GotResp:     &errs.Error{},
			ExpResp:     errs.Errorf(errs.InvalidArgument, errs.CodeValidation, "[{\"field\":\"order\",\"error\":\"unknown order: ser_id\"}]"),
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
			GotResp:     &errs.Error{},
			ExpResp:     errs.Errorf(errs.PermissionDenied, errs.CodePermissionDenied, "permission denied"),
			AssertFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "audit-query-403")
}
