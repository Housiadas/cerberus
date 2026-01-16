package apitest

import (
	"bytes"
	"context"
	"testing"

	"github.com/Housiadas/cerberus/internal/app/handler"
	"github.com/Housiadas/cerberus/internal/common/dbtest"
	cfg "github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/otel"
	"github.com/stretchr/testify/require"
)

// StartTest initialized the system to run a test.
func StartTest(t *testing.T, testName string) (*Test, error) {
	t.Helper()

	// Initialize test database
	db := dbtest.New(t, testName)

	// Initialize logger
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", "", "")

	// Initialize tracer
	traceProvider, teardown, err := otel.InitTracing(otel.Config{
		ServiceName: "Service Name",
		Host:        "Test host",
		ExcludedRoutes: map[string]struct{}{
			"/liveness":  {},
			"/readiness": {},
		},
		Probability: 0.5,
	})
	defer teardown(context.Background())

	require.NoError(t, err)

	tracer := traceProvider.Tracer("Service Name")

	// Initialize handler
	h := handler.New(handler.Config{
		ServiceName: "Test Service Name",
		Build:       "Test",
		Cors:        cfg.CorsSettings{},
		DB:          db,
		Log:         log,
		Tracer:      tracer,
	})

	// initialize apitest services
	c := newCore(log, db)

	// initialize usecase
	u := Usecase{Auth: h.Usecase.Auth}

	return New(db, h.Routes(), c, u), nil
}
