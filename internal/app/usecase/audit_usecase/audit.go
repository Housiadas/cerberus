// Package audit_usecase maintains the app layer api for the audit domain.
package audit_usecase

import (
	"context"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/product_core"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
)

type App struct {
	AuditCore *audit_core.Core
}

func NewApp(core *audit_core.Core) *App {
	return &App{
		AuditCore: core,
	}
}

func (a *App) Query(ctx context.Context, qp AppQueryParams) (page.Result[Audit], error) {
	p, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return page.Result[Audit]{}, validation.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Result[Audit]{}, err.(*errs.Error)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, user.DefaultOrderBy)
	if err != nil {
		return page.Result[Audit]{}, validation.NewFieldErrors("order", err)
	}

	adts, err := a.AuditCore.Query(ctx, filter, orderBy, p)
	if err != nil {
		return page.Result[Audit]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.AuditCore.Count(ctx, filter)
	if err != nil {
		return page.Result[Audit]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewResult(toAppAudits(adts), total, p), nil
}
