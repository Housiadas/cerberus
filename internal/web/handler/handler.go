//go:generate go tool oapi-codegen --config ../../../openapi/oapi-codegen.yaml ../../../openapi/openapi.yaml

package handler

import (
	"context"
	"time"

	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/internal/core/account/account_repo"
	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/audit/audit_repo"
	"github.com/Housiadas/cerberus/internal/core/auth"
	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox"
	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox/email_notification_outbox_repo"
	"github.com/Housiadas/cerberus/internal/core/outbox"
	"github.com/Housiadas/cerberus/internal/core/outbox/outbox_repo"
	"github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/core/permission/permission_cache"
	"github.com/Housiadas/cerberus/internal/core/permission/permission_repo"
	"github.com/Housiadas/cerberus/internal/core/refresh_token"
	"github.com/Housiadas/cerberus/internal/core/refresh_token/refresh_token_repo"
	"github.com/Housiadas/cerberus/internal/core/reset_token"
	"github.com/Housiadas/cerberus/internal/core/reset_token/reset_token_repo"
	"github.com/Housiadas/cerberus/internal/core/role"
	"github.com/Housiadas/cerberus/internal/core/role/role_cache"
	"github.com/Housiadas/cerberus/internal/core/role/role_repo"
	"github.com/Housiadas/cerberus/internal/core/role_permissions"
	"github.com/Housiadas/cerberus/internal/core/role_permissions/role_permissions_repo"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/core/user/user_cache"
	"github.com/Housiadas/cerberus/internal/core/user/user_repo"
	"github.com/Housiadas/cerberus/internal/core/user_roles"
	"github.com/Housiadas/cerberus/internal/core/user_roles/user_roles_repo"
	"github.com/Housiadas/cerberus/internal/core/user_roles_permissions"
	"github.com/Housiadas/cerberus/internal/core/user_roles_permissions/user_roles_permissions_repo"
	"github.com/Housiadas/cerberus/internal/sdk/eventbus"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/internal/web/middleware"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Ensure Handler implements the strict server interface at compile time.
var _ openapi.StrictServerInterface = (*Handler)(nil)

// Client defines the interface for Redis operations used by distributed storage.
type redisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
	Pipeline() redis.Pipeliner
}

// Config represents the configuration for the Handler.
type Config struct {
	ServiceName       string
	Build             string
	Cors              config.CorsSettings
	Stripe            config.Stripe
	DB                *sqlx.DB
	Redis             redisClient
	Log               logger.Logger
	Tracer            trace.Tracer
	Meter             metric.Meter
	AccessTokenSecret []byte
	FrontendURL       string
}

// Handler contains all the mandatory systems required by Handler.
type Handler struct {
	serviceName string
	build       string
	db          *sqlx.DB
	log         logger.Logger
	cors        config.CorsSettings
	middleware  *middleware.Middleware
	svc         services
}

// services holds direct references to all domain services.
type services struct {
	auth                 *auth.Service
	user                 *user.Service
	role                 *role.Service
	permission           *permission.Service
	audit                *audit.Service
	userRolesPermissions *user_roles_permissions.Service
	userRoles            *user_roles.Service
	rolePermissions      *role_permissions.Service
	account              *account.Service
}

func New(ctx context.Context, cfg Config) *Handler {
	// utils
	hash := hasher.NewBcrypt()
	clk := clock.NewClock()
	uuidGen := uuidgen.NewV7()

	// http clients
	//stripeClient := stripepkg.New(stripepkg.Config{
	//	WebhookSecret: cfg.Stripe.WebhookSecret,
	//	SecretKey:     cfg.Stripe.SecretKey,
	//	Log:           cfg.Log,
	//})

	// repos
	auditRepo := audit_repo.NewStore(cfg.Log, cfg.DB)
	outboxRepo := outbox_repo.NewStore(cfg.Log, cfg.DB)
	emailNotifOutboxRepo := email_notification_outbox_repo.NewStore(cfg.Log, cfg.DB)
	userRepo := user_repo.NewStore(cfg.Log, cfg.DB)
	userCacheStore := user_cache.NewStore(ctx, cfg.Log, userRepo, cfg.Redis)
	roleRepo := role_repo.NewStore(cfg.Log, cfg.DB)
	roleCacheStore := role_cache.NewStore(ctx, cfg.Log, roleRepo, cfg.Redis)
	permissionRepo := permission_repo.NewStore(cfg.Log, cfg.DB)
	permissionCacheStore := permission_cache.NewStore(ctx, cfg.Log, permissionRepo, cfg.Redis)
	userRolesPermissionsRepo := user_roles_permissions_repo.NewStore(cfg.Log, cfg.DB)
	userRolesRepo := user_roles_repo.NewStore(cfg.Log, cfg.DB, clk)
	rolePermissionsRepo := role_permissions_repo.NewStore(cfg.Log, cfg.DB, clk)
	refreshTokenRepo := refresh_token_repo.NewStore(cfg.Log, cfg.DB)
	resetTokenRepo := reset_token_repo.NewStore(cfg.Log, cfg.DB)
	accountRepo := account_repo.NewStore(cfg.Log, cfg.DB)

	// services
	auditService := audit.NewService(cfg.Log, auditRepo, clk)
	outboxSvc := outbox.NewService(cfg.Log, outboxRepo, uuidGen, clk)
	emailNotifOutboxSvc := email_notification_outbox.NewService(
		cfg.Log,
		emailNotifOutboxRepo,
		uuidGen,
		clk,
	)

	// event dispatcher and transaction beginner (used by services for CUD operations)
	dispatcher := eventbus.New(outboxSvc, auditService)
	tx := pgsql.NewTransactor(cfg.Log, cfg.DB)

	userService := user.NewService(cfg.Log, userCacheStore, uuidGen, clk, hash, tx, dispatcher)
	roleService := role.NewService(cfg.Log, roleCacheStore, uuidGen, tx, dispatcher, clk)
	permissionService := permission.NewService(
		cfg.Log,
		permissionCacheStore,
		uuidGen,
		tx,
		dispatcher,
		clk,
	)
	refreshTokenService := refresh_token.NewService(cfg.Log, refreshTokenRepo, uuidGen, clk)
	resetTokenService := reset_token.NewService(resetTokenRepo, uuidGen, clk)
	userRolesSvc := user_roles.NewService(cfg.Log, userRolesRepo, tx, dispatcher)
	rolePermsSvc := role_permissions.NewService(cfg.Log, rolePermissionsRepo, tx, dispatcher)
	userRolesPermissionsService := user_roles_permissions.NewService(
		cfg.Log,
		userRolesPermissionsRepo,
	)
	accountSvc := account.NewService(cfg.Log, accountRepo, uuidGen, clk, tx)

	authService := auth.NewService(auth.Config{
		Issuer:                     cfg.ServiceName,
		AccessTokenSecret:          cfg.AccessTokenSecret,
		Clock:                      clk,
		Log:                        cfg.Log,
		UserService:                userService,
		RefreshTokenService:        refreshTokenService,
		ResetTokenService:          resetTokenService,
		EmailNotificationOutboxSvc: emailNotifOutboxSvc,
		TX:                         tx,
		FrontendURL:                cfg.FrontendURL,
		UserRolesPermissions:       userRolesPermissionsService,
	})

	return &Handler{
		serviceName: cfg.ServiceName,
		build:       cfg.Build,
		db:          cfg.DB,
		log:         cfg.Log,
		cors:        cfg.Cors,
		middleware: middleware.New(ctx, middleware.Config{
			Log:                  cfg.Log,
			Tracer:               cfg.Tracer,
			Meter:                cfg.Meter,
			UserService:          userService,
			AuthUseCase:          authService,
			UserRolesPermissions: userRolesPermissionsService,
		}),
		svc: services{
			audit:                auditService,
			auth:                 authService,
			user:                 userService,
			role:                 roleService,
			permission:           permissionService,
			userRolesPermissions: userRolesPermissionsService,
			userRoles:            userRolesSvc,
			rolePermissions:      rolePermsSvc,
			account:              accountSvc,
		},
	}
}
