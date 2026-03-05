package apitest

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/app/handler"
	"github.com/Housiadas/cerberus/internal/app/relay"
	cfg "github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/internal/utils/dbtest"
	"github.com/Housiadas/cerberus/internal/utils/kafkatest"
	"github.com/Housiadas/cerberus/internal/utils/redistest"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	"github.com/stretchr/testify/require"
)

// StartTest initialized the system to run a test.
func StartTest(t *testing.T, testName string) (*Test, error) {
	t.Helper()

	// Initialize test database
	db := dbtest.New(t, testName)

	// Initialize logger
	var buf bytes.Buffer

	traceIDFn := func(context.Context) string {
		return ""
	}
	requestIDFn := func(context.Context) string {
		return ""
	}
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDFn, requestIDFn)

	// Initialize tracer
	ctx := context.Background()

	tel, err := telemetry.New(ctx, telemetry.Config{
		ServiceName: "Service Name",
	})
	require.NoError(t, err)

	defer tel.Shutdown(context.Background()) //nolint:errcheck

	// Initialize Redis testcontainers
	red := redistest.New(t)

	// Initialize handler
	serviceName := "Test Service Name"
	accessTokenSecret := []byte("test-256-bit-access-secret")
	h := handler.New(ctx, handler.Config{
		ServiceName:       serviceName,
		Build:             "Test",
		Cors:              cfg.CorsSettings{},
		DB:                db,
		Redis:             red,
		Log:               log,
		Tracer:            tel.TracerProvider().Tracer("Service Name"),
		Meter:             tel.MeterProvider().Meter("Service Name"),
		AccessTokenSecret: accessTokenSecret,
	})

	// Initialize Kafka producer via testcontainers
	kafkaProducer := kafkatest.New(t)

	// initialize logic
	c := newCore(log, db)
	u := newUseCase(log, c, accessTokenSecret, serviceName)

	outboxRelay := relay.New(log, c.Outbox, kafkaProducer, 1*time.Second, 100, 5)

	relayCtx, relayCancel := context.WithCancel(ctx)

	go outboxRelay.Start(relayCtx)

	t.Cleanup(func() {
		relayCancel()
	})

	return New(db, h.Routes(), c, u), nil
}
