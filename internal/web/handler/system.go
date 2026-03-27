package handler

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

func (h *Handler) Readiness(
	ctx context.Context,
	_ openapi.ReadinessRequestObject,
) (openapi.ReadinessResponseObject, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := pgsql.StatusCheck(ctx, h.db)
	if err != nil {
		h.log.Info(ctx, "readiness failure", "ERROR", err)

		return nil, errs.Errorf(errs.Internal, errs.CodeInternal, "database not ready")
	}

	return openapi.Readiness200JSONResponse{
		Status: new("None"),
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
		Status:     new("up"),
		Build:      new(h.build),
		Host:       new(host),
		Name:       new(os.Getenv("KUBERNETES_NAME")),
		PodIp:      new(os.Getenv("KUBERNETES_POD_IP")),
		Node:       new(os.Getenv("KUBERNETES_NODE_NAME")),
		Namespace:  new(os.Getenv("KUBERNETES_NAMESPACE")),
		Gomaxprocs: new(runtime.GOMAXPROCS(0)),
	}, nil
}
