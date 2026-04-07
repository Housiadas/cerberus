package unitest

import (
	"context"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/core/role"
	"github.com/Housiadas/cerberus/internal/core/user"
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

// Permission represents a permission specified for the test.
type Permission struct {
	permission.Permission
}

// SeedData represents data seeded for the test.
type SeedData struct {
	Users       []User
	Admins      []User
	Roles       []Role
	Permissions []Permission
}

// The Table represents fields needed for running a unit test.
type Table struct {
	Name    string
	ExpResp any
	ExcFunc func(ctx context.Context) any
	CmpFunc func(got any, exp any) string
}
