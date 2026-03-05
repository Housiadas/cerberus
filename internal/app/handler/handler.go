//go:generate go tool oapi-codegen --config ../../../openapi/oapi-codegen.yaml ../../../openapi/openapi.yaml

package handler

import (
	"context"

	"github.com/Housiadas/cerberus/internal/app/cache/user_cache"
	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	"github.com/Housiadas/cerberus/internal/app/middleware"
	"github.com/Housiadas/cerberus/internal/app/repo/audit_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/outbox_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/permission_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/refresh_token_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/role_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/user_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/user_roles_permissions_repo"
	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/outbox_service"
	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/internal/core/service/refresh_token_service"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_permissions_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/internal/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/permission_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/refresh_token_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/role_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/system_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_roles_permissions_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/redis"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Ensure Handler implements the strict server interface at compile time.
var _ openapi.StrictServerInterface = (*Handler)(nil)

// Handler contains all the mandatory systems required by handler.
type Handler struct {
	ServiceName string
	Build       string
	Cors        config.CorsSettings
	DB          *sqlx.DB
	Log         logger.Logger
	Middleware  *middleware.Middleware
	Usecase     Usecase
}

// Usecase represents the use case layer.
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
	ServiceName       string
	Build             string
	Cors              config.CorsSettings
	DB                *sqlx.DB
	Redis             redis.Client
	Log               logger.Logger
	Tracer            trace.Tracer
	Meter             metric.Meter
	AccessTokenSecret []byte
}

func New(ctx context.Context, cfg Config) *Handler {
	// utils
	hash := hasher.NewBcrypt()
	clk := clock.NewClock()
	uuidGen := uuidgen.NewV7()

	// repos
	auditRepo := audit_repo.NewStore(cfg.Log, cfg.DB)
	outboxRepo := outbox_repo.NewStore(cfg.Log, cfg.DB)
	userRepo := user_repo.NewStore(cfg.Log, cfg.DB)
	userCacheStore := user_cache.NewStore(ctx, cfg.Log, userRepo, cfg.Redis)
	roleRepo := role_repo.NewStore(cfg.Log, cfg.DB)
	permissionRepo := permission_repo.NewStore(cfg.Log, cfg.DB)
	userRolesPermissionsRepo := user_roles_permissions_repo.NewStore(cfg.Log, cfg.DB)
	refreshTokenRepo := refresh_token_repo.NewStore(cfg.Log, cfg.DB)

	// services
	auditService := audit_service.New(cfg.Log, auditRepo)
	outboxSvc := outbox_service.New(cfg.Log, outboxRepo, uuidGen, clk)
	userService := user_service.New(cfg.Log, userCacheStore, uuidGen, clk, hash)
	roleService := role_service.New(cfg.Log, roleRepo)
	permissionService := permission_service.New(cfg.Log, permissionRepo)
	refreshTokenService := refresh_token_service.New(cfg.Log, refreshTokenRepo, uuidGen, clk)
	userRolesPermissionsService := user_roles_permissions_service.New(
		cfg.Log,
		userRolesPermissionsRepo,
	)

	// usecase
	auditUsecase := audit_usecase.NewUseCase(auditService)
	userUsecase := user_usecase.NewUseCase(userService, outboxSvc)
	refreshTokenUsecase := refresh_token_usecase.NewUseCase(refreshTokenService)
	authUsecase := auth_usecase.NewUseCase(auth_usecase.Config{
		Issuer:              cfg.ServiceName,
		AccessTokenSecret:   cfg.AccessTokenSecret,
		Log:                 cfg.Log,
		UserUsecase:         userUsecase,
		RefreshTokenUsecase: refreshTokenUsecase,
	})
	roleUsecase := role_usecase.NewUseCase(roleService)
	permissionUsecase := permission_usecase.NewUseCase(permissionService)
	systemUsecase := system_usecase.NewUseCase(cfg.Build, cfg.Log, cfg.DB)
	userRolesPermissionsUsecase := user_roles_permissions_usecase.NewUseCase(
		userRolesPermissionsService,
	)

	return &Handler{
		ServiceName: cfg.ServiceName,
		Build:       cfg.Build,
		Cors:        cfg.Cors,
		DB:          cfg.DB,
		Log:         cfg.Log,
		Middleware: middleware.New(ctx, middleware.Config{
			Log:                  cfg.Log,
			Tracer:               cfg.Tracer,
			Meter:                cfg.Meter,
			Tx:                   pgsql.NewBeginner(cfg.DB),
			UserUseCase:          userUsecase,
			AuthUseCase:          authUsecase,
			UserRolesPermissions: userRolesPermissionsUsecase,
		}),
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
