// Package audit_usecase maintains the use case layer api the model
package audit_usecase

import (
	"context"
	"errors"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/audit_service"
	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
)

type UseCase struct {
	AuditService *audit_service.Service
}

func NewUseCase(service *audit_service.Service) *UseCase {
	return &UseCase{
		AuditService: service,
	}
}

func (a *UseCase) Query(ctx context.Context, qp AppQueryParams) (cursor.Result[Audit], error) {
	cur, err := cursor.Parse(qp.Cursor, qp.Limit)
	if err != nil {
		return cursor.Result[Audit]{}, errs2.NewFieldErrors("cursor", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return cursor.Result[Audit]{}, func() *errs2.Error {
			target := &errs2.Error{}
			_ = errors.As(err, &target)

			return target
		}()
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, user.GetDefaultOrderBy())
	if err != nil {
		return cursor.Result[Audit]{}, errs2.NewFieldErrors("order", err)
	}

	adts, err := a.AuditService.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return cursor.Result[Audit]{}, errs2.Errorf(
			errs2.Internal,
			errs2.CodeInternal,
			"query: %s",
			err,
		)
	}

	return cursor.NewResult(
		toAppAudits(adts),
		cur.Limit(),
		cur,
		orderBy,
		func(a Audit) string { return a.ID },
		auditFieldExtractor(orderBy),
	), nil
}
