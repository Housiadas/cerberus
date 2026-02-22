package apitest

import (
	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/usecase/auth_usecase"
)

// User extends the dbtest user for api test support.
type User struct {
	user.User

	AccessToken auth_usecase.AccessToken
	Audits      []audit.Audit
}

// SeedData represents users for api tests.
type SeedData struct {
	Users  []User
	Admins []User
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
