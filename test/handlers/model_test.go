package handlers_test

import (
	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/core/role"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/sdk/jsonutil"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/pkg/clock"
)

func toTestAudit(a audit.Audit) openapi.Audit {
	return openapi.Audit{
		Id:        a.ID().String(),
		ObjId:     a.ObjID().String(),
		ObjEntity: a.ObjEntity().String(),
		ObjName:   a.ObjName().String(),
		ActorId:   a.ActorID().String(),
		Action:    a.Action(),
		Data:      jsonutil.Compact(a.Data()),
		Message:   a.Message(),
		Timestamp: clock.Format(new(a.Timestamp())),
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
		Id:        p.ID().String(),
		Name:      p.Name().String(),
		CreatedAt: clock.Format(new(p.CreatedAt())),
		UpdatedAt: clock.Format(new(p.UpdatedAt())),
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
		Id:        r.ID().String(),
		Name:      r.Name().String(),
		CreatedAt: clock.Format(new(r.CreatedAt())),
		UpdatedAt: clock.Format(new(r.UpdatedAt())),
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
		CreatedAt:  clock.Format(new(bus.CreatedAt())),
		UpdatedAt:  clock.Format(new(bus.UpdatedAt())),
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
