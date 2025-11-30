package dbtest

import (
	"github.com/jmoiron/sqlx"

	"github.com/Housiadas/cerberus/internal/app/repo/audit_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/role_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/user_repo"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/pkg/logger"
)

// Service represents all the internal core apis needed for testing.
type Service struct {
	Audit *audit_service.Service
	User  *user_service.Service
	Role  *role_service.Service
}

func newCore(log *logger.Logger, db *sqlx.DB) Service {
	auditService := audit_service.New(log, audit_repo.NewStore(log, db))
	userService := user_service.New(log, user_repo.NewStore(log, db))
	roleService := role_service.New(log, role_repo.NewStore(log, db))

	return Service{
		Audit: auditService,
		User:  userService,
		Role:  roleService,
	}
}
