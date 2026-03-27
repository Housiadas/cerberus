// Package apitest provides support for integration http tests.
package apitest

import (
	"net/http"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/auth"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/jmoiron/sqlx"
)

// Test contains functions for executing an integration test.
type Test struct {
	DB   *sqlx.DB
	Mux  http.Handler
	Auth *auth.Service
	Core *Core
}

// New constructs a Test value for running api tests.
func New(db *sqlx.DB, mux http.Handler, c *Core, authSvc *auth.Service) *Test {
	return &Test{
		DB:   db,
		Mux:  mux,
		Core: c,
		Auth: authSvc,
	}
}

// User extends the dbtest user for api test support.
type User struct {
	user.User

	AccessToken auth.AccessToken
	Audits      []audit.Audit
}

// Table represents fields needed for running an api test.
type Table struct {
	Name        string
	URL         string
	Method      string
	AccessToken *string
	StatusCode  int
	Input       any
	GotResp     any
	ExpResp     any
	AssertFunc  func(got any, exp any) string
}
