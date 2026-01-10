package handler

import (
	"context"
	"net/http"

	"github.com/Housiadas/cerberus/internal/app/usecase/system_usecase"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

// readiness checks if the database is ready and if not will return a 500 status.
// Do not respond by just returning an error because further up in the call
// stack it will interpret that as a non-trusted error.
//
// Readiness godoc
//
//	@Summary		Usecase Readiness
//	@Description	Check application's readiness
//	@Tags			System
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	system_usecase.Status
//	@Failure		500	{object}	errs.Error
//	@Router			/readiness [get].
func (h *Handler) readiness(
	ctx context.Context,
	_ http.ResponseWriter,
	_ *http.Request,
) web.Encoder {
	err := h.Usecase.System.Readiness(ctx)
	if err != nil {
		return errs.Errorf(errs.Internal, "database not ready")
	}

	data := system_usecase.Status{
		Status: "None",
	}

	return data
}

// liveness returns simple status info if the usecase is alive. If the
// cli is deployed to a Kubernetes cluster, it will also return pod, node, and
// namespace details via the Downward API. The Kubernetes environment variables
// need to be set within your Pod/Deployment manifest.
//
// Liveness godoc
//
//	@Summary		Usecase Liveness
//	@Description	Returns application's status info if the usecase is alive
//	@Tags			System
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	system_usecase.Info
//	@Router			/liveness [get].
func (h *Handler) liveness(_ context.Context, _ http.ResponseWriter, _ *http.Request) web.Encoder {
	info := h.Usecase.System.Liveness()

	return info
}
