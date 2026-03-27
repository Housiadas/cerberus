package handlers_test

import (
	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/core/role"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/pkg/clock"
)

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T { return &v }

func toTestAudit(a audit.Audit) openapi.Audit {
	return openapi.Audit{
		Id:        new(a.ID().String()),
		ObjId:     new(a.ObjID().String()),
		ObjEntity: new(a.ObjEntity().String()),
		ObjName:   new(a.ObjName().String()),
		ActorId:   new(a.ActorID().String()),
		Action:    new(a.Action()),
		Data:      new(string(a.Data())),
		Message:   new(a.Message()),
		Timestamp: new(clock.Format(ptr(a.Timestamp()))),
	}
}

func toTestAudits(audits []audit.Audit) []openapi.Audit {
	out := make([]openapi.Audit, len(audits))
	for i, a := range audits {
		out[i] = toTestAudit(a)
	}

	return out
}

func toTestPermission(p permission.Permission) openapi.Permission {
	return openapi.Permission{
		Id:        new(p.ID().String()),
		Name:      new(p.Name().String()),
		CreatedAt: new(clock.Format(ptr(p.CreatedAt()))),
		UpdatedAt: new(clock.Format(ptr(p.UpdatedAt()))),
	}
}

func toTestPermissions(perms []permission.Permission) []openapi.Permission {
	out := make([]openapi.Permission, len(perms))
	for i, p := range perms {
		out[i] = toTestPermission(p)
	}

	return out
}

func toTestRole(r role.Role) openapi.Role {
	return openapi.Role{
		Id:        new(r.ID().String()),
		Name:      new(r.Name().String()),
		CreatedAt: new(clock.Format(ptr(r.CreatedAt()))),
		UpdatedAt: new(clock.Format(ptr(r.UpdatedAt()))),
	}
}

func toTestRoles(roles []role.Role) []openapi.Role {
	out := make([]openapi.Role, len(roles))
	for i, r := range roles {
		out[i] = toTestRole(r)
	}

	return out
}

func toUserResponse(bus user.User) openapi.User {
	return openapi.User{
		Id:         bus.ID().String(),
		Name:       bus.Name().String(),
		Email:      bus.Email().Address,
		Department: bus.Department().String(),
		Enabled:    bus.Enabled(),
		CreatedAt:  clock.Format(ptr(bus.CreatedAt())),
		UpdatedAt:  clock.Format(ptr(bus.UpdatedAt())),
	}
}

func toUserResponses(users []user.User) []openapi.User {
	items := make([]openapi.User, len(users))
	for i, usr := range users {
		items[i] = toUserResponse(usr)
	}

	return items
}

func toUserResponsePtr(bus user.User) *openapi.User {
	resp := toUserResponse(bus)

	return &resp
}
