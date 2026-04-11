//go:generate go tool oapi-codegen --config ../../../openapi/oapi-codegen.yaml ../../../openapi/openapi.yaml

package handler

import (
	"context"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/auth"
	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox"
	"github.com/Housiadas/cerberus/internal/core/outbox"
	"github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/core/refresh_token"
	"github.com/Housiadas/cerberus/internal/core/reset_token"
	"github.com/Housiadas/cerberus/internal/core/role"
	"github.com/Housiadas/cerberus/internal/core/role_permissions"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/core/user/user_cache"
	"github.com/Housiadas/cerberus/internal/core/user_roles"
	"github.com/Housiadas/cerberus/internal/core/user_roles_permissions"
	"github.com/Housiadas/cerberus/internal/sdk/eventbus"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/internal/web/middleware"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/jackc/pgx/v5/pgxpool"
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
	DB                *pgxpool.Pool
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
	store       *db.Store
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

func New(cfg Config) *Handler {
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

	// db layer
	store := db.NewStore(cfg.DB)
	tx := db.NewTransactor(cfg.Log, cfg.DB)

	// cache layer
	userCache := user_cache.NewStore(cfg.Log, store, cfg.Redis)

	// services
	auditService := audit.NewService(cfg.Log, store, clk)
	outboxSvc := outbox.NewService(cfg.Log, store, uuidGen, clk)
	emailNotifOutboxSvc := email_notification_outbox.NewService(
		cfg.Log,
		store,
		uuidGen,
		clk,
	)

	// event dispatcher and transaction beginner (used by services for CUD operations)
	dispatcher := eventbus.New(outboxSvc, auditService)

	userService := user.NewService(cfg.Log, userCache, uuidGen, clk, hash, tx, dispatcher)
	roleService := role.NewService(cfg.Log, store, uuidGen, tx, dispatcher, clk)
	permissionService := permission.NewService(
		cfg.Log,
		store,
		uuidGen,
		tx,
		dispatcher,
		clk,
	)
	refreshTokenService := refresh_token.NewService(cfg.Log, store, uuidGen, clk)
	resetTokenService := reset_token.NewService(store, uuidGen, clk)
	userRolesSvc := user_roles.NewService(cfg.Log, store, tx, dispatcher)
	rolePermsSvc := role_permissions.NewService(cfg.Log, store, tx, dispatcher)
	userRolesPermissionsService := user_roles_permissions.NewService(
		cfg.Log,
		store,
	)
	accountSvc := account.NewService(cfg.Log, store, uuidGen, clk, tx)

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
		store:       store,
		log:         cfg.Log,
		cors:        cfg.Cors,
		middleware: middleware.New(middleware.Config{
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
