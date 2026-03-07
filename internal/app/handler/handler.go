//go:generate go tool oapi-codegen --config ../../../openapi/oapi-codegen.yaml ../../../openapi/openapi.yaml

package handler

import (
	"context"

	"github.com/Housiadas/cerberus/internal/app/cache/permission_cache"
	"github.com/Housiadas/cerberus/internal/app/cache/role_cache"
	"github.com/Housiadas/cerberus/internal/app/cache/user_cache"
	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	"github.com/Housiadas/cerberus/internal/app/middleware"
	"github.com/Housiadas/cerberus/internal/app/repo/audit_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/outbox_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/permission_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/refresh_token_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/role_permissions_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/role_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/user_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/user_roles_permissions_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/user_roles_repo"
	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/outbox_service"
	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/internal/core/service/refresh_token_service"
	"github.com/Housiadas/cerberus/internal/core/service/role_permissions_service"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_permissions_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/internal/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/permission_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/refresh_token_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/role_permissions_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/role_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/system_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_roles_permissions_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_roles_usecase"
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

// Config represents the configuration for the Handler.
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

// Handler contains all the mandatory systems required by Handler.
type Handler struct {
	serviceName string
	cors        config.CorsSettings
	middleware  *middleware.Middleware
	usecase     usecase
}

// usecase represents the use case layer.
type usecase struct {
	auth                 *auth_usecase.UseCase
	user                 *user_usecase.UseCase
	role                 *role_usecase.UseCase
	audit                *audit_usecase.UseCase
	permission           *permission_usecase.UseCase
	userRolesPermissions *user_roles_permissions_usecase.UseCase
	userRoles            *user_roles_usecase.UseCase
	rolePermissions      *role_permissions_usecase.UseCase
	system               *system_usecase.UseCase
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
	roleCacheStore := role_cache.NewStore(ctx, cfg.Log, roleRepo, cfg.Redis)
	permissionRepo := permission_repo.NewStore(cfg.Log, cfg.DB)
	permissionCacheStore := permission_cache.NewStore(ctx, cfg.Log, permissionRepo, cfg.Redis)
	userRolesPermissionsRepo := user_roles_permissions_repo.NewStore(cfg.Log, cfg.DB)
	userRolesRepo := user_roles_repo.NewStore(cfg.Log, cfg.DB)
	rolePermissionsRepo := role_permissions_repo.NewStore(cfg.Log, cfg.DB)
	refreshTokenRepo := refresh_token_repo.NewStore(cfg.Log, cfg.DB)

	// services
	auditService := audit_service.New(cfg.Log, auditRepo)
	outboxSvc := outbox_service.New(cfg.Log, outboxRepo, uuidGen, clk)
	userService := user_service.New(cfg.Log, userCacheStore, uuidGen, clk, hash)
	roleService := role_service.New(cfg.Log, roleCacheStore, uuidGen)
	permissionService := permission_service.New(cfg.Log, permissionCacheStore, uuidGen)
	refreshTokenService := refresh_token_service.New(cfg.Log, refreshTokenRepo, uuidGen, clk)
	userRolesSvc := user_roles_service.New(cfg.Log, userRolesRepo)
	rolePermsSvc := role_permissions_service.New(cfg.Log, rolePermissionsRepo)
	userRolesPermissionsService := user_roles_permissions_service.New(
		cfg.Log,
		userRolesPermissionsRepo,
	)

	// usecase
	tx := pgsql.NewBeginner(cfg.DB)
	auditUsecase := audit_usecase.NewUseCase(auditService)
	userUsecase := user_usecase.NewUseCase(cfg.Log, userService, outboxSvc, auditService, tx)
	refreshTokenUsecase := refresh_token_usecase.NewUseCase(refreshTokenService)
	authUsecase := auth_usecase.NewUseCase(auth_usecase.Config{
		Issuer:              cfg.ServiceName,
		AccessTokenSecret:   cfg.AccessTokenSecret,
		Log:                 cfg.Log,
		UserUsecase:         userUsecase,
		RefreshTokenUsecase: refreshTokenUsecase,
	})
	roleUsecase := role_usecase.NewUseCase(cfg.Log, roleService, auditService, tx)
	permissionUsecase := permission_usecase.NewUseCase(cfg.Log, permissionService, auditService, tx)
	systemUsecase := system_usecase.NewUseCase(cfg.Build, cfg.Log, cfg.DB)
	userRolesPermissionsUsecase := user_roles_permissions_usecase.NewUseCase(
		user_roles_permissions_usecase.Config{
			Service: userRolesPermissionsService,
		},
	)
	userRolesUsecase := user_roles_usecase.NewUseCase(cfg.Log, userRolesSvc, auditService, tx)
	rolePermissionsUsecase := role_permissions_usecase.NewUseCase(
		cfg.Log, rolePermsSvc, auditService, tx,
	)

	return &Handler{
		serviceName: cfg.ServiceName,
		cors:        cfg.Cors,
		middleware: middleware.New(ctx, middleware.Config{
			Log:                  cfg.Log,
			Tracer:               cfg.Tracer,
			Meter:                cfg.Meter,
			UserUseCase:          userUsecase,
			AuthUseCase:          authUsecase,
			UserRolesPermissions: userRolesPermissionsUsecase,
		}),
		usecase: usecase{
			audit:                auditUsecase,
			auth:                 authUsecase,
			user:                 userUsecase,
			role:                 roleUsecase,
			permission:           permissionUsecase,
			userRolesPermissions: userRolesPermissionsUsecase,
			userRoles:            userRolesUsecase,
			rolePermissions:      rolePermissionsUsecase,
			system:               systemUsecase,
		},
	}
}
