package apitest

import (
	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/audit/audit_repo"
	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox"
	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox/email_notification_outbox_repo"
	"github.com/Housiadas/cerberus/internal/core/outbox"
	"github.com/Housiadas/cerberus/internal/core/outbox/outbox_repo"
	"github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/core/permission/permission_repo"
	"github.com/Housiadas/cerberus/internal/core/refresh_token"
	"github.com/Housiadas/cerberus/internal/core/refresh_token/refresh_token_repo"
	"github.com/Housiadas/cerberus/internal/core/reset_token"
	"github.com/Housiadas/cerberus/internal/core/reset_token/reset_token_repo"
	"github.com/Housiadas/cerberus/internal/core/role"
	"github.com/Housiadas/cerberus/internal/core/role/role_repo"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/core/user/user_repo"
	"github.com/Housiadas/cerberus/internal/core/user_roles_permissions"
	"github.com/Housiadas/cerberus/internal/core/user_roles_permissions/user_roles_permissions_repo"
	"github.com/Housiadas/cerberus/internal/eventbus"
	"github.com/Housiadas/cerberus/internal/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/refresh_token_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_roles_permissions_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/jmoiron/sqlx"
)

type Dependency struct {
	Core    *Core
	Usecase *Usecase
}

// Core represents all the internal core services needed for testing.
type Core struct {
	Audit                   *audit.Service
	User                    *user.Service
	Role                    *role.Service
	Permission              *permission.Service
	RefreshToken            *refresh_token.Service
	Outbox                  *outbox.Service
	ResetToken              *reset_token.Service
	EmailNotificationOutbox *email_notification_outbox.Service
}

type Usecase struct {
	Auth *auth_usecase.UseCase
}

func newDependency(
	log *logger.Service,
	db *sqlx.DB,
	accessTokenSecret []byte,
	serviceName string,
) *Dependency {
	// utils
	clk := clock.NewClock()
	hash := hasher.NewBcrypt()
	uuidGen := uuidgen.NewV7()

	// services
	auditService := audit.NewService(log, audit_repo.NewStore(log, db))
	roleService := role.NewService(log, role_repo.NewStore(log, db), uuidGen)
	outboxSvc := outbox.NewService(log, outbox_repo.NewStore(log, db), uuidGen, clk)
	userService := user.NewService(log, user_repo.NewStore(log, db), uuidGen, clk, hash)
	permissionService := permission.NewService(log, permission_repo.NewStore(log, db), uuidGen)
	refreshTokenService := refresh_token.NewService(
		log,
		refresh_token_repo.NewStore(log, db),
		uuidGen,
		clk,
	)
	resetTokenSvc := reset_token.NewService(reset_token_repo.NewStore(log, db), uuidGen, clk)
	emailNotificationOutboxSvc := email_notification_outbox.NewService(
		log,
		email_notification_outbox_repo.NewStore(log, db),
		uuidGen,
		clk,
	)

	// event dispatcher
	dispatcher := eventbus.New(outboxSvc, auditService)

	// usecase
	userUsecase := user_usecase.NewUseCase(
		log,
		userService,
		dispatcher,
		pgsql.NewBeginner(db),
	)
	refreshTokenUsecase := refresh_token_usecase.NewUseCase(refreshTokenService)
	userRolesPermissionsSvc := user_roles_permissions.NewService(
		log,
		user_roles_permissions_repo.NewStore(log, db),
	)
	userRolesPermissionsUsecase := user_roles_permissions_usecase.NewUseCase(
		user_roles_permissions_usecase.Config{Service: userRolesPermissionsSvc},
	)
	authUsecase := auth_usecase.NewUseCase(auth_usecase.Config{
		Issuer:                     serviceName,
		AccessTokenSecret:          accessTokenSecret,
		Log:                        log,
		UserUsecase:                userUsecase,
		RefreshTokenUsecase:        refreshTokenUsecase,
		UserService:                userService,
		ResetTokenService:          resetTokenSvc,
		EmailNotificationOutboxSvc: emailNotificationOutboxSvc,
		DB:                         pgsql.NewBeginner(db),
		FrontendURL:                "http://localhost:3000",
		UserRolesPermissions:       userRolesPermissionsUsecase,
	})

	return &Dependency{
		Core: &Core{
			Audit:                   auditService,
			User:                    userService,
			RefreshToken:            refreshTokenService,
			Role:                    roleService,
			Permission:              permissionService,
			Outbox:                  outboxSvc,
			ResetToken:              resetTokenSvc,
			EmailNotificationOutbox: emailNotificationOutboxSvc,
		},
		Usecase: &Usecase{Auth: authUsecase},
	}
}
