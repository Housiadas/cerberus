package handler

import (
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"

	"github.com/Housiadas/cerberus/internal/app/middleware"
	"github.com/Housiadas/cerberus/internal/app/repo/audit_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/permission_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/role_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/user_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/user_roles_permissions_repo"
	"github.com/Housiadas/cerberus/internal/app/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/auth_usecase"
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
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/Housiadas/cerberus/pkg/web"
)

// Handler contains all the mandatory systems required by handler.
type Handler struct {
	ServiceName string
	Build       string
	Cors        config.CorsSettings
	DB          *sqlx.DB
	Log         logger.Logger
	Tracer      trace.Tracer
	Web         Web
	Usecase     Usecase
}

// Web represents the set of usecase for the http.
type Web struct {
	Middleware *middleware.Middleware
	Res        *web.Respond
}

// Usecase represents the use case layer
type Usecase struct {
	Audit                *audit_usecase.UseCase
	Auth                 *auth_usecase.UseCase
	User                 *user_usecase.UseCase
	Role                 *role_usecase.UseCase
	Permission           *permission_usecase.UseCase
	UserRolesPermissions *user_roles_permissions_usecase.UseCase
	System               *system_usecase.UseCase
}

// Config represents the configuration for the handler.
type Config struct {
	ServiceName string
	Build       string
	Cors        config.CorsSettings
	DB          *sqlx.DB
	Log         logger.Logger
	Tracer      trace.Tracer
}

func New(cfg Config) *Handler {
	// utils
	hash := hasher.NewBcrypt()
	clk := clock.NewClock()
	uuidGen := uuidgen.NewV7()

	// repos
	auditRepo := audit_repo.NewStore(cfg.Log, cfg.DB)
	userRepo := user_repo.NewStore(cfg.Log, cfg.DB)
	roleRepo := role_repo.NewStore(cfg.Log, cfg.DB)
	permissionRepo := permission_repo.NewStore(cfg.Log, cfg.DB)
	userRolesPermissionsRepo := user_roles_permissions_repo.NewStore(cfg.Log, cfg.DB)

	// services
	auditService := audit_service.New(cfg.Log, auditRepo)
	userService := user_service.New(cfg.Log, userRepo, uuidGen, clk, hash)
	roleService := role_service.New(cfg.Log, roleRepo)
	permissionService := permission_service.New(cfg.Log, permissionRepo)
	userRolesPermissionsService := user_roles_permissions_service.New(cfg.Log, userRolesPermissionsRepo)

	// usecase
	auditUsecase := audit_usecase.NewUseCase(auditService)
	userUsecase := user_usecase.NewUseCase(userService)
	authUsecase := auth_usecase.NewUseCase(auth_usecase.Config{
		Issuer:      cfg.ServiceName,
		Log:         cfg.Log,
		UserUsecase: userUsecase,
	})
	roleUsecase := role_usecase.NewUseCase(roleService)
	permissionUsecase := permission_usecase.NewUseCase(permissionService)
	systemUsecase := system_usecase.NewUseCase(cfg.Build, cfg.Log, cfg.DB)
	userRolesPermissionsUsecase := user_roles_permissions_usecase.NewUseCase(userRolesPermissionsService)

	return &Handler{
		ServiceName: cfg.ServiceName,
		Build:       cfg.Build,
		Cors:        cfg.Cors,
		DB:          cfg.DB,
		Log:         cfg.Log,
		Tracer:      cfg.Tracer,
		Web: Web{
			Middleware: middleware.New(middleware.Config{
				Log:                  cfg.Log,
				Tracer:               cfg.Tracer,
				Tx:                   pgsql.NewBeginner(cfg.DB),
				UserUseCase:          userUsecase,
				AuthUseCase:          authUsecase,
				UserRolesPermissions: userRolesPermissionsUsecase,
			}),
			Res: web.NewRespond(cfg.Log),
		},
		Usecase: Usecase{
			Audit:                auditUsecase,
			Auth:                 authUsecase,
			User:                 userUsecase,
			Role:                 roleUsecase,
			Permission:           permissionUsecase,
			UserRolesPermissions: userRolesPermissionsUsecase,
			System:               systemUsecase,
		},
	}
}
