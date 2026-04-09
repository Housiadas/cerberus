package apitest

import (
	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/auth"
	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox"
	"github.com/Housiadas/cerberus/internal/core/outbox"
	"github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/internal/core/refresh_token"
	"github.com/Housiadas/cerberus/internal/core/reset_token"
	"github.com/Housiadas/cerberus/internal/core/role"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/core/user_roles_permissions"
	"github.com/Housiadas/cerberus/internal/sdk/eventbus"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/hasher"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/jmoiron/sqlx"
)

type Dependency struct {
	Core *Core
	Auth *auth.Service
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

func newDependency(
	log logger.Logger,
	db *sqlx.DB,
	accessTokenSecret []byte,
	serviceName string,
) *Dependency {
	// utils
	clk := clock.NewClock()
	hash := hasher.NewBcrypt()
	uuidGen := uuidgen.NewV7()

	// services
	auditService := audit.NewService(log, audit_repo.NewStore(log, db), clk)
	outboxSvc := outbox.NewService(log, outbox_repo.NewStore(log, db), uuidGen, clk)

	// event dispatcher and transactor (used by services for CUD operations)
	dispatcher := eventbus.New(outboxSvc, auditService)
	tx := pgsql.NewTransactor(log, db)
	userService := user.NewService(
		log,
		user_repo.NewStore(log, db),
		uuidGen,
		clk,
		hash,
		tx,
		dispatcher,
	)
	roleService := role.NewService(log, role_repo.NewStore(log, db), uuidGen, tx, dispatcher, clk)
	permissionService := permission.NewService(
		log,
		permission_repo.NewStore(log, db),
		uuidGen,
		tx,
		dispatcher,
		clk,
	)
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

	userRolesPermissionsSvc := user_roles_permissions.NewService(
		log,
		user_roles_permissions_repo.NewStore(log, db),
	)

	authService := auth.NewService(auth.Config{
		Issuer:                     serviceName,
		AccessTokenSecret:          accessTokenSecret,
		Clock:                      clk,
		Log:                        log,
		UserService:                userService,
		RefreshTokenService:        refreshTokenService,
		ResetTokenService:          resetTokenSvc,
		EmailNotificationOutboxSvc: emailNotificationOutboxSvc,
		TX:                         tx,
		FrontendURL:                "http://localhost:3000",
		UserRolesPermissions:       userRolesPermissionsSvc,
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
		Auth: authService,
	}
}
