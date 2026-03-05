// Package apitest provides support for integration http tests.
package apitest

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

// Test contains functions for executing an api test.
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
