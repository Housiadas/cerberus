package handlers_test

import (
	"net/http"
	"sort"
	"strings"
	"testing"

	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/testutil/apitest"
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

	expData := toTestAudits(sd.Admins[0].Audits)
	expMd := openapi.Metadata{
		HasMore: new(false),
		Limit:   new(10),
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
				Data:     &expData,
			},
			AssertFunc: func(got any, exp any) string {
				gotResp, exists := got.(*openapi.AuditPageResult)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*openapi.AuditPageResult)

				if gotResp.Data != nil && expResp.Data != nil {
					for i := range *gotResp.Data {
						gotA := &(*gotResp.Data)[i]
						expA := &(*expResp.Data)[i]

						if gotA.Timestamp != nil && expA.Timestamp != nil && *gotA.Timestamp == *expA.Timestamp {
							expA.Timestamp = gotA.Timestamp
						}

						if gotA.Data != nil {
							cleaned := strings.ReplaceAll(*gotA.Data, " ", "")
							gotA.Data = &cleaned
						}

						if expA.Data != nil {
							cleaned := strings.ReplaceAll(*expA.Data, " ", "")
							expA.Data = &cleaned
						}
					}
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
