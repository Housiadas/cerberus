// Package audit_usecase maintains the use case layer api the model
package audit_usecase

import (
	"context"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

type UseCase struct {
	AuditService *audit_service.Service
}

func NewUseCase(service *audit_service.Service) *UseCase {
	return &UseCase{
		AuditService: service,
	}
}

func (a *UseCase) Query(ctx context.Context, qp AppQueryParams) (web.Result[Audit], error) {
	p, err := web.Parse(qp.Page, qp.Rows)
	if err != nil {
		return web.Result[Audit]{}, errs.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return web.Result[Audit]{}, err.(*errs.Error)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, user.DefaultOrderBy)
	if err != nil {
		return web.Result[Audit]{}, errs.NewFieldErrors("order", err)
	}

	adts, err := a.AuditService.Query(ctx, filter, orderBy, p)
	if err != nil {
		return web.Result[Audit]{}, errs.Errorf(errs.Internal, "query: %s", err)
	}

	total, err := a.AuditService.Count(ctx, filter)
	if err != nil {
		return web.Result[Audit]{}, errs.Errorf(errs.Internal, "count: %s", err)
	}

	return web.NewResult(toAppAudits(adts), total, p), nil
}
