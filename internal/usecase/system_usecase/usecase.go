// Package system_usecase maintains the usecase layer http for the check core.
package system_usecase

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/jmoiron/sqlx"
)

type logger interface {
	Info(ctx context.Context, msg string, args ...any)
}

type UseCase struct {
	build string
	log   logger
	db    *sqlx.DB
}

func NewUseCase(build string, log logger, db *sqlx.DB) *UseCase {
	return &UseCase{
		build: build,
		log:   log,
		db:    db,
	}
}

// Readiness checks if the database is ready and if not will return a 500 status.
// Do not respond by just returning an error because further up in the call
// stack it will interpret that as a non-trusted error.
func (a *UseCase) Readiness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := pgsql.StatusCheck(ctx, a.db)
	if err != nil {
		a.log.Info(ctx, "readiness failure", "ERROR", err)

		return errs.New(errs.Internal, errs.CodeInternal, err)
	}

	return nil
}

// Liveness returns simple status info if the usecase is alive.
// If the app is deployed to a Kubernetes cluster, it will also return pod, node, and
// namespace details via the Downward API. The Kubernetes environment variables
// need to be set within your Pod/Deployment manifest.
func (a *UseCase) Liveness() Info {
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	info := Info{
		Status:     "up",
		Build:      a.build,
		Host:       host,
		Name:       os.Getenv("KUBERNETES_NAME"),
		PodIP:      os.Getenv("KUBERNETES_POD_IP"),
		Node:       os.Getenv("KUBERNETES_NODE_NAME"),
		Namespace:  os.Getenv("KUBERNETES_NAMESPACE"),
		GOMAXPROCS: runtime.GOMAXPROCS(0),
	}

	// This handler provides a free timer loop.

	return info
}
