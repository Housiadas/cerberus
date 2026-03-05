package handler

import (
	"context"

	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	"github.com/Housiadas/cerberus/internal/usecase/system_usecase"
	"github.com/Housiadas/cerberus/internal/utils/errs"
)

func (h *Handler) Readiness(
	ctx context.Context,
	_ openapi.ReadinessRequestObject,
) (openapi.ReadinessResponseObject, error) {
	err := h.usecase.system.Readiness(ctx)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, "database not ready")
	}

	return openapi.Readiness200JSONResponse(system_usecase.Status{Status: "None"}), nil
}

func (h *Handler) Liveness(
	_ context.Context,
	_ openapi.LivenessRequestObject,
) (openapi.LivenessResponseObject, error) {
	return openapi.Liveness200JSONResponse(h.usecase.system.Liveness()), nil
}
