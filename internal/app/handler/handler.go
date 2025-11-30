package handler

import (
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"

	"github.com/Housiadas/cerberus/internal/app/middleware"
	"github.com/Housiadas/cerberus/internal/app/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/role_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/system_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
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
	Core        Core
}

// Web represents the set of usecase for the http.
type Web struct {
	Middleware *middleware.Middleware
	Res        *web.Respond
}

// UseCase represents the application layer
type UseCase struct {
	Audit  *audit_usecase.UseCase
	User   *user_usecase.UseCase
	Role   *role_usecase.UserCase
	System *system_usecase.UseCase
}

// Core represents the core internal layer.
type Core struct {
	Audit *audit_service.Service
	User  *user_service.Service
}

// Config represents the configuration for the handler.
type Config struct {
	ServiceName  string
	Build        string
	Cors         config.CorsSettings
	DB           *sqlx.DB
	Log          *logger.Logger
	Tracer       trace.Tracer
	AuditService *audit_service.Service
	UserService  *user_service.Service
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
			Audit:  audit_usecase.NewApp(cfg.AuditService),
			User:   user_usecase.NewApp(cfg.UserService),
			System: system_usecase.NewApp(cfg.Build, cfg.Log, cfg.DB),
		},
		Core: Core{
			Audit: cfg.AuditService,
			User:  cfg.UserService,
		},
	}
}
