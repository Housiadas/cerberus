package handler

import (
	"context"

	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	"github.com/Housiadas/cerberus/internal/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/internal/utils/pntr"
)

func (h *Handler) ListAudits(
	ctx context.Context,
	request openapi.ListAuditsRequestObject,
) (openapi.ListAuditsResponseObject, error) {
	qp := audit_usecase.AppQueryParams{
		Page:      pntr.DerefStr(request.Params.Page),
		Rows:      pntr.DerefStr(request.Params.Rows),
		OrderBy:   pntr.DerefStr(request.Params.OrderBy),
		ObjID:     pntr.DerefStr(request.Params.ObjId),
		ObjEntity: pntr.DerefStr(request.Params.ObjDomain),
		ObjName:   pntr.DerefStr(request.Params.ObjName),
		ActorID:   pntr.DerefStr(request.Params.ActorId),
		Action:    pntr.DerefStr(request.Params.Action),
		Since:     pntr.DerefStr(request.Params.Since),
		Until:     pntr.DerefStr(request.Params.Until),
	}

	result, err := h.Usecase.Audit.Query(ctx, qp)
	if err != nil {
		return nil, err
	}

	return openapi.ListAudits200JSONResponse{
		Data:     new(result.Data),
		Metadata: new(result.Metadata),
	}, nil
}
