package apitest

import (
	"github.com/Housiadas/cerberus/internal/app/repo/audit_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/role_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/user_repo"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/jmoiron/sqlx"
)

func newCore(log *logger.Service, db *sqlx.DB) Core {
	// utils
	hash := hasher.NewBcrypt()
	clk := clock.NewClock()
	uuidGen := uuidgen.NewV7()
	// services
	auditService := audit_service.New(log, audit_repo.NewStore(log, db))
	userService := user_service.New(log, user_repo.NewStore(log, db), uuidGen, clk, hash)
	roleService := role_service.New(log, role_repo.NewStore(log, db))

	return Core{
		Audit: auditService,
		User:  userService,
		Role:  roleService,
	}
}
