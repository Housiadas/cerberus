package handler

import (
	"context"

	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/usecase/system_usecase"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
)

func (h *Handler) Readiness(
	ctx context.Context,
	_ openapi.ReadinessRequestObject,
) (openapi.ReadinessResponseObject, error) {
	err := h.usecase.system.Readiness(ctx)
	if err != nil {
		return nil, errs2.Errorf(errs2.Internal, errs2.CodeInternal, "database not ready")
	}

	return openapi.Readiness200JSONResponse(system_usecase.Status{Status: "None"}), nil
}

func (h *Handler) Liveness(
	_ context.Context,
	_ openapi.LivenessRequestObject,
) (openapi.LivenessResponseObject, error) {
	return openapi.Liveness200JSONResponse(h.usecase.system.Liveness()), nil
}
