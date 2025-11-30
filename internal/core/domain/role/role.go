// Package role represents the role type in the system.
package role

import (
	"errors"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("role not found")
)

// Role represents information about an individual role of our system.
type Role struct {
	ID          uuid.UUID
	Name        name.Name
	DateCreated time.Time
	DateUpdated time.Time
}

type NewRole struct {
	Name name.Name
}

type UpdateRole struct {
	Name *name.Name
}

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID   *uuid.UUID
	Name *name.Name
}

//// The set of roles that can be used.
//var (
//	Admin = newRole("ADMIN")
//	User  = newRole("USER")
//)
//
//// Set of known roles.
//var roles = make(map[string]Role)
//
//// Role represents a role in the system.
//type Role struct {
//	value string
//}
//
//func newRole(role string) Role {
//	r := Role{role}
//	roles[role] = r
//	return r
//}
//
//// String returns the name of the role.
//func (r Role) String() string {
//	return r.value
//}
//
//// Equal provides support for the go-cmp package and testing.
//func (r Role) Equal(r2 Role) bool {
//	return r.value == r2.value
//}
//
//// =============================================================================
//
//// Parse parses the string value and returns a role if one exists.
//func Parse(value string) (Role, error) {
//	role, exists := roles[value]
//	if !exists {
//		return Role{}, fmt.Errorf("invalid role %q", value)
//	}
//
//	return role, nil
//}
//
//// MustParse parses the string value and returns a role if one exists. If
//// an error occurs, the function panics.
//func MustParse(value string) Role {
//	role, err := Parse(value)
//	if err != nil {
//		panic(err)
//	}
//
//	return role
//}
//
//// ParseToString takes a collection of user roles and converts them to
//// a slice of string.
//func ParseToString(usrRoles []Role) []string {
//	roles := make([]string, len(usrRoles))
//	for i, role := range usrRoles {
//		roles[i] = role.String()
//	}
//
//	return roles
//}
//
//// ParseMany takes a collection of strings and converts them to a slice
//// of roles.
//func ParseMany(roles []string) ([]Role, error) {
//	usrRoles := make([]Role, len(roles))
//	for i, roleStr := range roles {
//		role, err := Parse(roleStr)
//		if err != nil {
//			return nil, err
//		}
//		usrRoles[i] = role
//	}
//
//	return usrRoles, nil
//}
