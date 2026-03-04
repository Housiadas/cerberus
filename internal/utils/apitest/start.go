package apitest

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/Housiadas/cerberus/internal/app/handler"
	"github.com/Housiadas/cerberus/internal/app/relay"
	"github.com/Housiadas/cerberus/internal/app/repo/outbox_repo"
	cfg "github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/internal/core/service/outbox_service"
	"github.com/Housiadas/cerberus/internal/utils/dbtest"
	"github.com/Housiadas/cerberus/internal/utils/kafkatest"
	"github.com/Housiadas/cerberus/internal/utils/redistest"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
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

	tracer := tel.TracerProvider().Tracer("Service Name")

	// Initialize Redis testcontainers
	red := redistest.New(t)

	// Initialize handler
	h := handler.New(ctx, handler.Config{
		ServiceName:       "Test Service Name",
		Build:             "Test",
		Cors:              cfg.CorsSettings{},
		DB:                db,
		Redis:             red,
		Log:               log,
		Tracer:            tracer,
		AccessTokenSecret: []byte("test-256-bit-access-secret"),
	})

	// Initialize Kafka producer via testcontainers
	kafkaProducer := kafkatest.New(t)

	// Start outbox relay
	outboxRepo := outbox_repo.NewStore(log, db)
	outboxSvc := outbox_service.New(log, outboxRepo, uuidgen.NewV7(), clock.NewClock())
	outboxRelay := relay.New(log, outboxSvc, kafkaProducer, 1*time.Second, 100, 5)

	relayCtx, relayCancel := context.WithCancel(ctx)

	go outboxRelay.Start(relayCtx)

	// initialize apitest services
	c := newCore(log, db)

	// initialize usecase
	u := Usecase{Auth: h.Usecase.Auth}

	t.Cleanup(func() {
		relayCancel()
	})

	return New(db, h.Routes(), c, u), nil
}
