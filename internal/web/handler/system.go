package handler

import (
	"context"
	"os"
	"runtime"
	"time"

	errs "github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
)

func (h *Handler) Readiness(
	ctx context.Context,
	_ openapi.ReadinessRequestObject,
) (openapi.ReadinessResponseObject, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := h.store.Status(ctx)
	if err != nil {
		h.log.Error(ctx, "readiness failure", "ERROR", err)

		return nil, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"database not ready",
		)
	}

	return openapi.Readiness200JSONResponse{
		Status: "Ready",
	}, nil
}

func (h *Handler) Liveness(
	_ context.Context,
	_ openapi.LivenessRequestObject,
) (openapi.LivenessResponseObject, error) {
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	return openapi.Liveness200JSONResponse{
		Status:     "up",
		Build:      h.build,
		Host:       host,
		Name:       os.Getenv("KUBERNETES_NAME"),
		PodIp:      os.Getenv("KUBERNETES_POD_IP"),
		Node:       os.Getenv("KUBERNETES_NODE_NAME"),
		Namespace:  os.Getenv("KUBERNETES_NAMESPACE"),
		Gomaxprocs: runtime.GOMAXPROCS(0),
	}, nil
}
