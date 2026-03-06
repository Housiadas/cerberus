package handlers_test

import (
	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/permission_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/role_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/clock"
)

func toAppAudit(bus audit.Audit) audit_usecase.Audit {
	return audit_usecase.Audit{
		ID:        bus.ID().String(),
		ObjID:     bus.ObjID().String(),
		ObjEntity: bus.ObjEntity().String(),
		ObjName:   bus.ObjName().String(),
		ActorID:   bus.ActorID().String(),
		Action:    bus.Action(),
		Data:      string(bus.Data()),
		Message:   bus.Message(),
		Timestamp: clock.Format(new(bus.Timestamp())),
	}
}

func toAppAudits(audits []audit.Audit) []audit_usecase.Audit {
	app := make([]audit_usecase.Audit, len(audits))
	for i, adt := range audits {
		app[i] = toAppAudit(adt)
	}

	return app
}

func toAppPermission(p permission.Permission) permission_usecase.Permission {
	return permission_usecase.Permission{
		ID:        p.ID().String(),
		Name:      p.Name().String(),
		CreatedAt: clock.Format(new(p.CreatedAt())),
		UpdatedAt: clock.Format(new(p.UpdatedAt())),
	}
}

func toAppPermissions(perms []permission.Permission) []permission_usecase.Permission {
	items := make([]permission_usecase.Permission, len(perms))
	for i, p := range perms {
		items[i] = toAppPermission(p)
	}

	return items
}

func toAppRole(r role.Role) role_usecase.Role {
	return role_usecase.Role{
		ID:        r.ID().String(),
		Name:      r.Name().String(),
		CreatedAt: clock.Format(new(r.CreatedAt())),
		UpdatedAt: clock.Format(new(r.UpdatedAt())),
	}
}

func toAppRoles(roles []role.Role) []role_usecase.Role {
	items := make([]role_usecase.Role, len(roles))
	for i, r := range roles {
		items[i] = toAppRole(r)
	}

	return items
}

func toAppUser(bus user.User) user_usecase.User {
	return user_usecase.User{
		ID:           bus.ID().String(),
		Name:         bus.Name().String(),
		Email:        bus.Email().Address,
		PasswordHash: nil,
		Department:   bus.Department().String(),
		Enabled:      bus.Enabled(),
		CreatedAt:    clock.Format(new(bus.CreatedAt())),
		UpdatedAt:    clock.Format(new(bus.UpdatedAt())),
	}
}

func toAppUsers(users []user.User) []user_usecase.User {
	items := make([]user_usecase.User, len(users))
	for i, usr := range users {
		items[i] = toAppUser(usr)
	}

	return items
}

func toAppUserPtr(bus user.User) *user_usecase.User {
	return new(toAppUser(bus))
}
