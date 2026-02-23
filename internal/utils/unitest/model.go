package unitest

import (
	"context"

	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
)

// User represents a user specified for the test.
type User struct {
	user.User

	Audits []audit.Audit
}

// Role represents a role specified for the test.
type Role struct {
	role.Role
}

// SeedData represents data seeded for the test.
type SeedData struct {
	Users  []User
	Admins []User
	Roles  []Role
}

// The Table represents fields needed for running a unit test.
type Table struct {
	Name    string
	ExpResp any
	ExcFunc func(ctx context.Context) any
	CmpFunc func(got any, exp any) string
}
