// Package apitest provides support for integration http tests.
package apitest

import (
	"net/http"

	"github.com/Housiadas/cerberus/internal/app/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/jmoiron/sqlx"
)

// Test contains functions for executing an api test.
type Test struct {
	DB      *sqlx.DB
	Mux     http.Handler
	Usecase Usecase
	Core    Core
}

// New constructs a Test value for running api tests.
func New(db *sqlx.DB, mux http.Handler, c Core, u Usecase) *Test {
	return &Test{
		DB:      db,
		Mux:     mux,
		Core:    c,
		Usecase: u,
	}
}

type Usecase struct {
	Auth *auth_usecase.UseCase
}

// Core represents all the internal core services needed for testing.
type Core struct {
	Audit *audit_service.Service
	User  *user_service.Service
	Role  *role_service.Service
}
