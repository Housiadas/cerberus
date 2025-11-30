package dbtest

import (
	"github.com/jmoiron/sqlx"

	"github.com/Housiadas/cerberus/internal/app/repo/audit_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/product_repo"
	"github.com/Housiadas/cerberus/internal/app/repo/user_repo"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/internal/core/service/product_core"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/pkg/logger"
)

// Core represents all the internal core apis needed for testing.
type Core struct {
	Audit   *audit_service.Core
	User    *user_service.Core
	Product *product_core.Core
}

func newCore(log *logger.Logger, db *sqlx.DB) Core {
	auditCore := audit_service.NewCore(log, audit_repo.NewStore(log, db))
	userBus := user_service.NewCore(log, user_repo.NewStore(log, db))
	productBus := product_core.NewCore(log, userBus, product_repo.NewStore(log, db))

	return Core{
		Audit:   auditCore,
		User:    userBus,
		Product: productBus,
	}
}
