// Package unitest provides support for executing unit test logic.
package unitest

import (
	"context"
	"net/mail"
	"testing"
)

// Run performs the actual test logic based on the table data.
func Run(t *testing.T, table []Table, testName string) {
	t.Helper()

	for _, tt := range table {
		t.Run(testName+"-"+tt.Name, cmpTest(tt))
	}
}

func cmpTest(tt Table) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		gotResp := tt.ExcFunc(context.Background())

		diff := tt.CmpFunc(gotResp, tt.ExpResp)
		if diff != "" {
			t.Log("DIFF")
			t.Logf("%s", diff)
			t.Log("GOT")
			t.Logf("%#v", gotResp)
			t.Log("EXP")
			t.Logf("%#v", tt.ExpResp)
			t.Fatalf("Should get the expected response")
		}
	}
}

func MustParseEmail(addr string) mail.Address {
	email, err := mail.ParseAddress(addr)
	if err != nil {
		panic(err)
	}

	return *email
}
