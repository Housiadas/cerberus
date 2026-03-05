// Package apitest provides support for integration http tests.
package apitest

import (
	"net/http"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/usecase/auth_usecase"
	"github.com/jmoiron/sqlx"
)

// Test contains functions for executing an integration test.
type Test struct {
	DB      *sqlx.DB
	Mux     http.Handler
	Usecase *Usecase
	Core    *Core
}

// New constructs a Test value for running api tests.
func New(db *sqlx.DB, mux http.Handler, c *Core, u *Usecase) *Test {
	return &Test{
		DB:      db,
		Mux:     mux,
		Core:    c,
		Usecase: u,
	}
}

// User extends the dbtest user for api test support.
type User struct {
	user.User

	AccessToken auth_usecase.AccessToken
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
