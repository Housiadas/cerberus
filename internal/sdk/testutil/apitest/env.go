package apitest

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	cfg "github.com/Housiadas/cerberus/internal/config"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/dbtest"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/kafkatest"
	"github.com/Housiadas/cerberus/internal/sdk/testutil/redistest"
	"github.com/Housiadas/cerberus/internal/web/handler"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/redis/go-redis/v9"
)

type producer interface {
	Produce(ctx context.Context, msg *ckafka.Message) error
	Flush(timeoutMs int) int
	Close()
}

// Client defines the interface for Redis operations used by distributed storage.
type redisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
	Pipeline() redis.Pipeliner
}

// Env holds shared containers started once for an entire test package via TestMain.
// Use SetupEnv in TestMain and call StartTest per test function.
type Env struct {
	pgShared      *dbtest.SharedContainer
	redisClient   redisClient
	kafkaProducer producer
}

// SetupEnv starts all required containers at once.
// Call it from a TestMain.
func SetupEnv() (*Env, func()) {
	ctx := context.Background()
	t := &mainT{}

	pg := dbtest.NewSharedContainer(ctx, t)
	red := redistest.NewSharedContainer(ctx, t)
	kfk := kafkatest.NewSharedContainer(ctx, t)

	env := &Env{
		pgShared:      pg,
		redisClient:   red,
		kafkaProducer: kfk,
	}

	teardown := func() {
		for i := len(t.cleanups) - 1; i >= 0; i-- {
			t.cleanups[i]()
		}
	}

	return env, teardown
}

// mainT is a minimal testing-like facade usable from TestMain (no *testing.M.Cleanup).
type mainT struct {
	cleanups []func()
}

func (m *mainT) Helper() {}

func (m *mainT) Fatal(args ...any) {
	panic(fmt.Sprint(args...))
}

func (m *mainT) Logf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func (m *mainT) Cleanup(fn func()) {
	m.cleanups = append(m.cleanups, fn)
}

// StartTest creates a fresh isolated database within the shared Postgres container,
// wires up the full handler stack, and returns a ready-to-use *Test.
func (e *Env) StartTest(t *testing.T, testName string) (*Test, error) {
	t.Helper()

	// Fresh DB for this test (drops on t.Cleanup)
	db := e.pgShared.NewDB(t)

	// Logger
	var buf bytes.Buffer

	log := logger.New(&buf, logger.LevelInfo, "TEST",
		func(context.Context) string { return "" },
		func(context.Context) string { return "" },
	)

	// Noop tracer/meter
	ctx := context.Background()

	tel, err := telemetry.New(ctx, telemetry.Config{ServiceName: "Test Service Name"})
	if err != nil {
		return nil, fmt.Errorf("telemetry.New: %w", err)
	}

	defer tel.Shutdown(context.Background()) //nolint:errcheck

	// Build handler
	serviceName := "Test Service Name"
	accessTokenSecret := []byte("test-256-bit-access-secret")
	h := handler.New(handler.Config{
		ServiceName:       serviceName,
		Build:             "Test",
		Cors:              cfg.CorsSettings{},
		DB:                db,
		Redis:             e.redisClient,
		Log:               log,
		Tracer:            tel.TracerProvider().Tracer(serviceName),
		Meter:             tel.MeterProvider().Meter(serviceName),
		AccessTokenSecret: accessTokenSecret,
	})

	// dependency injection
	dep := newDependency(log, db, accessTokenSecret, serviceName)

	t.Cleanup(func() {
		t.Logf(
			"*** LOGS (%s) ***\n\n%s\n*** LOGS (%s) ***\n",
			testName,
			buf.String(),
			testName,
		)
	})

	return New(db, h.Routes(), dep.Core, dep.Auth), nil
}
