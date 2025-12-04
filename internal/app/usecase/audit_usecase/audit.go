// Package audit_usecase maintains the use case layer api the model
package audit_usecase

import (
	"context"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
)

type UseCase struct {
	AuditService *audit_service.Service
}

func NewUseCase(service *audit_service.Service) *UseCase {
	return &UseCase{
		AuditService: service,
	}
}

func (a *UseCase) Query(ctx context.Context, qp AppQueryParams) (page.Result[Audit], error) {
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

	adts, err := a.AuditService.Query(ctx, filter, orderBy, p)
	if err != nil {
		return page.Result[Audit]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.AuditService.Count(ctx, filter)
	if err != nil {
		return page.Result[Audit]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return page.NewResult(toAppAudits(adts), total, p), nil
}
