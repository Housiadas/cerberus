package audit_test

import (
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/Housiadas/cerberus/internal/app/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/internal/common/apitest"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

func Test_API_Audit_Query_200(t *testing.T) {
	t.Parallel()

	test, err := apitest.StartTest(t, "Test_API_Audit")
	require.NoError(t, err)

	sd, err := insertSeedData(test)
	require.NoError(t, err)

	sort.Slice(sd.Users[0].Audits, func(i, j int) bool {
		return sd.Users[0].Audits[i].ObjName.String() <= sd.Users[0].Audits[j].ObjName.String()
	})

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/api/v1/audits?page=1&rows=10&orderBy=obj_name,ASC&obj_name=ObjName",
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &web.Result[audit_usecase.Audit]{},
			ExpResp: &web.Result[audit_usecase.Audit]{
				Metadata: web.Metadata{
					FirstPage:   1,
					CurrentPage: 1,
					LastPage:    1,
					RowsPerPage: 10,
					Total:       len(sd.Admins[0].Audits),
				},
				Data: toAppAudits(sd.Admins[0].Audits),
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*web.Result[audit_usecase.Audit])
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*web.Result[audit_usecase.Audit])

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

	test, err := apitest.StartTest(t, "Test_API_Audit")
	require.NoError(t, err)

	_, err = insertSeedData(test)
	require.NoError(t, err)

	table := []apitest.Table{
		{
			Name:       "bad-query-filter",
			URL:        "/api/v1/audits?page=1&rows=10&obj_id=123",
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Errorf(errs.InvalidArgument, "[{\"field\":\"obj_id\",\"error\":\"invalid UUID length: 3\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-order-by-value",
			URL:        "/api/v1/audits?page=1&rows=10&orderBy=ser_id,ASC",
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Errorf(errs.InvalidArgument, "[{\"field\":\"order\",\"error\":\"unknown order: ser_id\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	test.Run(t, table, "audit-query-400")
}
