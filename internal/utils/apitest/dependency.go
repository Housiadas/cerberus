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
	"github.com/Housiadas/cerberus/internal/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/refresh_token_usecase"
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
	Audit        *audit_service.Service
	User         *user_service.Service
	Role         *role_service.Service
	Permission   *permission_service.Service
	RefreshToken *refresh_token_service.Service
	Outbox       *outbox_service.Service
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

	// usecases
	userUsecase := user_usecase.NewUseCase(userService, outboxSvc, pgsql.NewBeginner(db))
	refreshTokenUsecase := refresh_token_usecase.NewUseCase(refreshTokenService)
	authUsecase := auth_usecase.NewUseCase(auth_usecase.Config{
		Issuer:              serviceName,
		AccessTokenSecret:   accessTokenSecret,
		Log:                 log,
		UserUsecase:         userUsecase,
		RefreshTokenUsecase: refreshTokenUsecase,
	})

	return &Dependency{
		Core: &Core{
			Audit:        auditService,
			User:         userService,
			RefreshToken: refreshTokenService,
			Role:         roleService,
			Permission:   permissionService,
			Outbox:       outboxSvc,
		},
		Usecase: &Usecase{Auth: authUsecase},
	}
}
