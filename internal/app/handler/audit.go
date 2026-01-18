package handler

import (
	"context"
	"net/http"

	"github.com/Housiadas/cerberus/internal/app/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

// Audit godoc
//
//	@Summary		Query Audits
//	@Description	Search audits in database based on criteria
//	@Tags			Audit
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	audit_usecase.AuditPageResult
//	@Failure		500	{object}	errs.Error
//	@Router			/v1/audits [get].
func (h *Handler) auditQuery(
	ctx context.Context,
	_ http.ResponseWriter,
	r *http.Request,
) web.Encoder {
	qp := auditParseQueryParams(r)

	audits, err := h.Usecase.Audit.Query(ctx, qp)
	if err != nil {
		return errs.AsErr(err)
	}

	return audits
}

func auditParseQueryParams(r *http.Request) audit_usecase.AppQueryParams {
	values := r.URL.Query()

	return audit_usecase.AppQueryParams{
		Page:      values.Get("page"),
		Rows:      values.Get("rows"),
		OrderBy:   values.Get("orderBy"),
		ObjID:     values.Get("obj_id"),
		ObjEntity: values.Get("obj_domain"),
		ObjName:   values.Get("obj_name"),
		ActorID:   values.Get("actor_id"),
		Action:    values.Get("action"),
		Since:     values.Get("since"),
		Until:     values.Get("until"),
	}
}
