package apitest

import (
	"context"
	"fmt"
	"testing"

	"github.com/Housiadas/cerberus/internal/app/handler"
	"github.com/Housiadas/cerberus/internal/common/dbtest"
	cfg "github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/pkg/otel"
)

// StartTest initialized the system to run a test.
func StartTest(t *testing.T, testName string) (*Test, error) {
	db := dbtest.New(t, testName)

	// tracer
	traceProvider, teardown, err := otel.InitTracing(otel.Config{
		Log:         db.Log,
		ServiceName: "Service Name",
		Host:        "Test host",
		ExcludedRoutes: map[string]struct{}{
			"/v1/liveness":  {},
			"/v1/readiness": {},
		},
		Probability: 0.5,
	})
	if err != nil {
		return nil, fmt.Errorf("starting tracing: %w", err)
	}

	defer teardown(context.Background())

	tracer := traceProvider.Tracer("Service Name")

	// Initialize handler
	h := handler.New(handler.Config{
		ServiceName: "Test Service Name",
		Build:       "Test",
		Cors:        cfg.CorsSettings{},
		DB:          db.DB,
		Log:         db.Log,
		Tracer:      tracer,
		AuditCore:   db.Core.Audit,
		UserCore:    db.Core.User,
	})

	return New(db, h.Routes()), nil
}
