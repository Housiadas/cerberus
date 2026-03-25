//go:generate go tool oapi-codegen --config ../../../openapi/oapi-codegen.yaml ../../../openapi/openapi.yaml

package handler

import (
	"context"

	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/internal/core/domain/account/account_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/audit/audit_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/email_notification_outbox/email_notification_outbox_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/invoice/invoice_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/outbox/outbox_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/permission/permission_cache"
	"github.com/Housiadas/cerberus/internal/core/domain/permission/permission_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/refresh_token/refresh_token_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/reset_token/reset_token_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/role/role_cache"
	"github.com/Housiadas/cerberus/internal/core/domain/role/role_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/role_permissions/role_permissions_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/subscription/subscription_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/user/user_cache"
	"github.com/Housiadas/cerberus/internal/core/domain/user/user_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/user_roles/user_roles_repo"
	"github.com/Housiadas/cerberus/internal/core/domain/user_roles_permissions/user_roles_permissions_repo"
	"github.com/Housiadas/cerberus/internal/core/service/account_service"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/email_notification_outbox_service"
	"github.com/Housiadas/cerberus/internal/core/service/invoice_service"
	"github.com/Housiadas/cerberus/internal/core/service/outbox_service"
	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/internal/core/service/refresh_token_service"
	"github.com/Housiadas/cerberus/internal/core/service/reset_token_service"
	"github.com/Housiadas/cerberus/internal/core/service/role_permissions_service"
	"github.com/Housiadas/cerberus/internal/core/service/role_service"
	"github.com/Housiadas/cerberus/internal/core/service/subscription_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_permissions_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_roles_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/internal/eventbus"
	"github.com/Housiadas/cerberus/internal/usecase/account_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/billing_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/permission_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/refresh_token_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/role_permissions_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/role_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/system_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_roles_permissions_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_roles_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_usecase"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/internal/web/middleware"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/redis"
	stripepkg "github.com/Housiadas/cerberus/pkg/stripe"
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
	Stripe            config.Stripe
	DB                *sqlx.DB
	Redis             redis.Client
	Log               logger.Logger
	Tracer            trace.Tracer
	Meter             metric.Meter
	AccessTokenSecret []byte
	FrontendURL       string
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
	account              *account_usecase.UseCase
	billing              *billing_usecase.UseCase
}

func New(ctx context.Context, cfg Config) *Handler {
	// utils
	hash := hasher.NewBcrypt()
	clk := clock.NewClock()
	uuidGen := uuidgen.NewV7()

	// http clients
	stripeClient := stripepkg.New(stripepkg.Config{
		WebhookSecret: cfg.Stripe.WebhookSecret,
		SecretKey:     cfg.Stripe.SecretKey,
		Log:           cfg.Log,
	})

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
	userRolesRepo := user_roles_repo.NewStore(cfg.Log, cfg.DB)
	rolePermissionsRepo := role_permissions_repo.NewStore(cfg.Log, cfg.DB)
	refreshTokenRepo := refresh_token_repo.NewStore(cfg.Log, cfg.DB)
	resetTokenRepo := reset_token_repo.NewStore(cfg.Log, cfg.DB)
	accountRepo := account_repo.NewStore(cfg.Log, cfg.DB)
	subscriptionRepo := subscription_repo.NewStore(cfg.Log, cfg.DB)
	invoiceRepo := invoice_repo.NewStore(cfg.Log, cfg.DB)

	// services
	auditService := audit_service.New(cfg.Log, auditRepo)
	outboxSvc := outbox_service.New(cfg.Log, outboxRepo, uuidGen, clk)
	emailNotifOutboxSvc := email_notification_outbox_service.New(
		cfg.Log, emailNotifOutboxRepo, uuidGen, clk,
	)
	userService := user_service.New(cfg.Log, userCacheStore, uuidGen, clk, hash)
	roleService := role_service.New(cfg.Log, roleCacheStore, uuidGen)
	permissionService := permission_service.New(cfg.Log, permissionCacheStore, uuidGen)
	refreshTokenService := refresh_token_service.New(cfg.Log, refreshTokenRepo, uuidGen, clk)
	resetTokenService := reset_token_service.New(cfg.Log, resetTokenRepo, uuidGen, clk)
	userRolesSvc := user_roles_service.New(cfg.Log, userRolesRepo)
	rolePermsSvc := role_permissions_service.New(cfg.Log, rolePermissionsRepo)
	userRolesPermissionsService := user_roles_permissions_service.New(
		cfg.Log,
		userRolesPermissionsRepo,
	)
	accountSvc := account_service.New(cfg.Log, accountRepo, uuidGen, clk)
	subscriptionSvc := subscription_service.New(cfg.Log, subscriptionRepo)
	invoiceSvc := invoice_service.New(cfg.Log, invoiceRepo)

	// event dispatcher
	dispatcher := eventbus.New(outboxSvc, auditService)

	// usecase
	tx := pgsql.NewBeginner(cfg.DB)
	auditUsecase := audit_usecase.NewUseCase(auditService)
	userUsecase := user_usecase.NewUseCase(cfg.Log, userService, dispatcher, tx)
	refreshTokenUsecase := refresh_token_usecase.NewUseCase(refreshTokenService)
	userRolesPermissionsUsecase := user_roles_permissions_usecase.NewUseCase(
		user_roles_permissions_usecase.Config{
			Service: userRolesPermissionsService,
		},
	)
	authUsecase := auth_usecase.NewUseCase(auth_usecase.Config{
		Issuer:                     cfg.ServiceName,
		AccessTokenSecret:          cfg.AccessTokenSecret,
		Log:                        cfg.Log,
		UserUsecase:                userUsecase,
		UserService:                userService,
		RefreshTokenUsecase:        refreshTokenUsecase,
		ResetTokenService:          resetTokenService,
		EmailNotificationOutboxSvc: emailNotifOutboxSvc,
		DB:                         tx,
		FrontendURL:                cfg.FrontendURL,
		UserRolesPermissions:       userRolesPermissionsUsecase,
	})
	roleUsecase := role_usecase.NewUseCase(cfg.Log, roleService, dispatcher, tx)
	permissionUsecase := permission_usecase.NewUseCase(cfg.Log, permissionService, dispatcher, tx)
	systemUsecase := system_usecase.NewUseCase(cfg.Build, cfg.Log, cfg.DB)
	userRolesUsecase := user_roles_usecase.NewUseCase(cfg.Log, userRolesSvc, dispatcher, tx)
	rolePermissionsUsecase := role_permissions_usecase.NewUseCase(
		cfg.Log, rolePermsSvc, dispatcher, tx,
	)
	accountUsecase := account_usecase.NewUseCase(cfg.Log, accountSvc, tx)
	billingUsecase := billing_usecase.NewUseCase(
		stripeClient, accountSvc, subscriptionSvc, invoiceSvc, uuidGen, clk,
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
			account:              accountUsecase,
			billing:              billingUsecase,
		},
	}
}
