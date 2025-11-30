package handler

import (
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"

	"github.com/Housiadas/cerberus/internal/app/middleware"
	"github.com/Housiadas/cerberus/internal/app/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/product_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/system_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/transaction_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/product_core"
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
	App         App
	Core        Core
}

// Web represents the set of usecase for the http.
type Web struct {
	Middleware *middleware.Middleware
	Res        *web.Respond
}

// App represents the application layer
type App struct {
	Audit   *audit_usecase.App
	User    *user_usecase.App
	Product *product_usecase.App
	System  *system_usecase.App
	Tx      *transaction_usecase.App
}

// Core represents the core internal layer.
type Core struct {
	Audit   *audit_service.Service
	User    *user_service.Service
	Product *product_core.Core
}

// Config represents the configuration for the handler.
type Config struct {
	ServiceName string
	Build       string
	Cors        config.CorsSettings
	DB          *sqlx.DB
	Log         *logger.Logger
	Tracer      trace.Tracer
	AuditCore   *audit_service.Service
	UserCore    *user_service.Service
	ProductCore *product_core.Core
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
				Log:     cfg.Log,
				Tracer:  cfg.Tracer,
				Tx:      pgsql.NewBeginner(cfg.DB),
				User:    cfg.UserCore,
				Product: cfg.ProductCore,
			}),
			Res: web.NewRespond(cfg.Log),
		},
		App: App{
			Audit:   audit_usecase.NewApp(cfg.AuditCore),
			User:    user_usecase.NewApp(cfg.UserCore),
			Product: product_usecase.NewApp(cfg.ProductCore),
			System:  system_usecase.NewApp(cfg.Build, cfg.Log, cfg.DB),
			Tx:      transaction_usecase.NewApp(cfg.UserCore, cfg.ProductCore),
		},
		Core: Core{
			Audit:   cfg.AuditCore,
			User:    cfg.UserCore,
			Product: cfg.ProductCore,
		},
	}
}
