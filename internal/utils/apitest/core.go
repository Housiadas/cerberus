package apitest

import (
	"github.com/Housiadas/cerberus/internal/app/repo/audit_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/outbox_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/permission_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/refresh_token_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/role_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/user_repo"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/outbox_service"
	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/internal/core/service/refresh_token_service"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/jmoiron/sqlx"
)

// Core represents all the internal core services needed for testing.
type Core struct {
	Audit        *audit_service.Service
	User         *user_service.Service
	Role         *role_service.Service
	Permission   *permission_service.Service
	RefreshToken *refresh_token_service.Service
	Outbox       *outbox_service.Service
}

func newCore(log *logger.Service, db *sqlx.DB) *Core {
	// utils
	clk := clock.NewClock()
	hash := hasher.NewBcrypt()
	uuidGen := uuidgen.NewV7()
	// services
	auditService := audit_service.New(log, audit_repo.NewStore(log, db))
	roleService := role_service.New(log, role_repo.NewStore(log, db), uuidGen)
	outboxSvc := outbox_service.New(log, outbox_repo.NewStore(log, db), uuidGen, clk)
	userService := user_service.New(log, user_repo.NewStore(log, db), uuidGen, clk, hash)
	permissionService := permission_service.New(log, permission_repo.NewStore(log, db), uuidGen)
	refreshTokenService := refresh_token_service.New(
		log,
		refresh_token_repo.NewStore(log, db),
		uuidGen,
		clk,
	)

	return &Core{
		Audit:        auditService,
		User:         userService,
		RefreshToken: refreshTokenService,
		Role:         roleService,
		Permission:   permissionService,
		Outbox:       outboxSvc,
	}
}
