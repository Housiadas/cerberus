package dbtest

import (
	"github.com/jmoiron/sqlx"

	"github.com/Housiadas/cerberus/internal/app/repository/audit_repo"
	"github.com/Housiadas/cerberus/internal/app/repository/product_repo"
	"github.com/Housiadas/cerberus/internal/app/repository/user_repo"
	"github.com/Housiadas/cerberus/internal/core/service/audit_core"
	"github.com/Housiadas/cerberus/internal/core/service/product_core"
	"github.com/Housiadas/cerberus/internal/core/service/user_core"
	"github.com/Housiadas/cerberus/pkg/logger"
)

// Core represents all the internal core apis needed for testing.
type Core struct {
	Audit   *audit_core.Core
	User    *user_core.Core
	Product *product_core.Core
}

func newCore(log *logger.Logger, db *sqlx.DB) Core {
	auditCore := audit_core.NewCore(log, audit_repo.NewStore(log, db))
	userBus := user_core.NewCore(log, user_repo.NewStore(log, db))
	productBus := product_core.NewCore(log, userBus, product_repo.NewStore(log, db))

	return Core{
		Audit:   auditCore,
		User:    userBus,
		Product: productBus,
	}
}
