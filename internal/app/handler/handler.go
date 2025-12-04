package handler

import (
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"

	"github.com/Housiadas/cerberus/internal/app/middleware"
	"github.com/Housiadas/cerberus/internal/app/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/permission_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/role_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/system_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/user_roles_permissions_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_permissions_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/web"
)

// Handler contains all the mandatory systems required by handler.
type Handler struct {
	ServiceName string
	Build       string
	Cors        config.CorsSettings
	DB          *sqlx.DB
	Log         *logger.Logger
	Tracer      trace.Tracer
	Web         Web
	UseCase     UseCase
}

// Web represents the set of usecase for the http.
type Web struct {
	Middleware *middleware.Middleware
	Res        *web.Respond
}

// UseCase represents the use case layer
type UseCase struct {
	Audit                *audit_usecase.UseCase
	User                 *user_usecase.UseCase
	Role                 *role_usecase.UseCase
	Permission           *permission_usecase.UseCase
	UserRolesPermissions *user_roles_permissions_usecase.UseCase
	System               *system_usecase.UseCase
}

// Config represents the configuration for the handler.
type Config struct {
	ServiceName                 string
	Build                       string
	Cors                        config.CorsSettings
	DB                          *sqlx.DB
	Log                         *logger.Logger
	Tracer                      trace.Tracer
	AuditService                *audit_service.Service
	UserService                 *user_service.Service
	RoleService                 *role_service.Service
	PermissionService           *permission_service.Service
	UserRolesPermissionsService *user_roles_permissions_service.Service
}

func New(cfg Config) *Handler {
	return &Handler{
		ServiceName: cfg.ServiceName,
		Build:       cfg.Build,
		Cors:        cfg.Cors,
		DB:          cfg.DB,
		Log:         cfg.Log,
		Tracer:      cfg.Tracer,
		Web: Web{
			Middleware: middleware.New(middleware.Config{
				Log:    cfg.Log,
				Tracer: cfg.Tracer,
				Tx:     pgsql.NewBeginner(cfg.DB),
				User:   cfg.UserService,
			}),
			Res: web.NewRespond(cfg.Log),
		},
		UseCase: UseCase{
			Audit:                audit_usecase.NewUseCase(cfg.AuditService),
			User:                 user_usecase.NewUseCase(cfg.UserService),
			Role:                 role_usecase.NewUseCase(cfg.RoleService),
			Permission:           permission_usecase.NewUseCase(cfg.PermissionService),
			UserRolesPermissions: user_roles_permissions_usecase.NewUseCase(cfg.UserRolesPermissionsService),
			System:               system_usecase.NewUseCase(cfg.Build, cfg.Log, cfg.DB),
		},
	}
}
