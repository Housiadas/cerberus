package handler

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	"github.com/Housiadas/cerberus/internal/usecase/audit_usecase"
	"github.com/Housiadas/cerberus/internal/utils/pntr"
)

func (h *Handler) ListAudits(
	ctx context.Context,
	request openapi.ListAuditsRequestObject,
) (openapi.ListAuditsResponseObject, error) {
	qp := audit_usecase.AppQueryParams{
		Cursor:    pntr.DerefStr(request.Params.Cursor),
		Limit:     pntr.DerefStr(request.Params.Limit),
		OrderBy:   pntr.DerefStr(request.Params.OrderBy),
		ObjID:     pntr.DerefStr(request.Params.ObjId),
		ObjEntity: pntr.DerefStr(request.Params.ObjDomain),
		ObjName:   pntr.DerefStr(request.Params.ObjName),
		ActorID:   pntr.DerefStr(request.Params.ActorId),
		Action:    pntr.DerefStr(request.Params.Action),
		Since:     pntr.DerefStr(request.Params.Since),
		Until:     pntr.DerefStr(request.Params.Until),
	}

	result, err := h.usecase.audit.Query(ctx, qp)
	if err != nil {
		return nil, fmt.Errorf("list audits: %w", err)
	}

	return openapi.ListAudits200JSONResponse{
		Data:     new(result.Data),
		Metadata: new(result.Metadata),
	}, nil
}
