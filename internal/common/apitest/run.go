package apitest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Run performs the actual test logic based on the table data.
func (at *Test) Run(t *testing.T, table []Table, testName string) {
	t.Helper()
	for _, tt := range table {
		t.Run(testName+"-"+tt.Name, at.server(tt))
	}
}

func (at *Test) server(tt Table) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		r := httptest.NewRequest(tt.Method, tt.URL, nil)
		w := httptest.NewRecorder()

		if tt.Input != nil {
			d, err := json.Marshal(tt.Input)
			if err != nil {
				t.Fatalf("Should be able to marshal the model : %s", err)
			}

			r = httptest.NewRequest(tt.Method, tt.URL, bytes.NewBuffer(d))
		}

		// add authorization JWT
		if tt.AccessToken != nil {
			r.Header.Set("Authorization", "Bearer "+*tt.AccessToken)
		}

		at.Mux.ServeHTTP(w, r)

		if w.Code != tt.StatusCode {
			t.Fatalf("%s: Should receive a status code of %d for the response : %d",
				tt.Name, tt.StatusCode, w.Code,
			)
		}

		if tt.StatusCode == http.StatusNoContent {
			return
		}

		err := json.Unmarshal(w.Body.Bytes(), tt.GotResp)
		if err != nil {
			t.Fatalf("Should be able to unmarshal the response : %s", err)
		}

		diff := tt.AssertFunc(tt.GotResp, tt.ExpResp)
		if diff != "" {
			t.Log("DIFF")
			t.Logf("%s", diff)
			t.Log("GOT")
			t.Logf("%#v", tt.GotResp)
			t.Log("EXP")
			t.Logf("%#v", tt.ExpResp)
			t.Fatalf("Should get the expected response")
		}
	}
}
