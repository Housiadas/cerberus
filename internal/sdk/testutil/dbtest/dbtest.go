package dbtest

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// SetupSharedContainer starts a Postgres container for use in TestMain.
// Returns the SharedContainer and a teardown function to call.
func SetupSharedContainer(ctx context.Context) (*SharedContainer, func()) {
	t := &mainT{}
	sc := NewSharedContainer(ctx, t)
	teardown := func() {
		for i := len(t.cleanups) - 1; i >= 0; i-- {
			t.cleanups[i]()
		}
	}

	return sc, teardown
}

// mainT is a minimal testing-like facade usable from TestMain.
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

// tHelper is the minimal interface required for shared container helpers.
type tHelper interface {
	Helper()
	Fatal(args ...any)
	Cleanup(fn func())
	Logf(format string, args ...any)
}

// SharedContainer holds a running Postgres container shared across tests in a package.
type SharedContainer struct {
	adminURL string
	cfg      Config
}

// NewSharedContainer starts a single Postgres container intended to be shared
// across all tests in a package via TestMain.
func NewSharedContainer(ctx context.Context, t tHelper) *SharedContainer {
	t.Helper()

	cfg := newConfigDirect(t)

	ctr, err := postgres.Run(
		ctx,
		cfg.PostgresImage,
		postgres.WithDatabase(cfg.DBName),
		postgres.WithUsername(cfg.DBUser),
		postgres.WithPassword(cfg.DBPassword),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	if err != nil {
		t.Fatal("start shared postgres container:", err)
	}

	t.Cleanup(func() {
		err := testcontainers.TerminateContainer(ctr)
		if err != nil {
			t.Logf("terminate shared postgres container: %s", err)
		}
	})

	adminURL, err := ctr.ConnectionString(ctx)
	if err != nil {
		t.Fatal("shared postgres connection string:", err)
	}

	return &SharedContainer{
		adminURL: adminURL,
		cfg:      cfg,
	}
}

// NewDB creates a fresh database within the shared container, runs migrations,
// and returns. The database is dropped on t.Cleanup.
func (sc *SharedContainer) NewDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	// Create the database
	adminDB, err := sql.Open("pgx", sc.adminURL+"sslmode=disable")
	require.NoError(t, err)

	dbName := testDBName()

	_, err = adminDB.ExecContext(ctx, "CREATE DATABASE "+dbName)
	if err != nil {
		t.Fatalf("create database %s: %s", dbName, err)
	}

	// Build URL for the new database
	testURL := replaceDBName(sc.adminURL, sc.cfg.DBName, dbName)

	// Run migrations
	err = migration(testURL)
	require.NoError(t, err)

	// Open connection
	dbConnUrl, err := pgxpool.ParseConfig(testURL)
	if err != nil {
		t.Fatalf("parse db config url %s: %s", dbName, err)
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, dbConnUrl)
	if err != nil {
		t.Fatalf("create db pool %s: %s", dbName, err)
	}

	t.Cleanup(func() {
		dbPool.Close()

		// adminDB at the end
		defer adminDB.Close()

		// Force-disconnect all clients then drop
		_, _ = adminDB.ExecContext(
			context.Background(),
			fmt.Sprintf(
				`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='%s'`,
				dbName,
			),
		)
		_, _ = adminDB.ExecContext(context.Background(), "DROP DATABASE IF EXISTS "+dbName)
	})

	return dbPool
}

func testDBName() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, 4)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}

// replaceDBName swaps the database name component in a Postgres URL.
func replaceDBName(url, oldName, newName string) string {
	if strings.HasPrefix(url, "postgres://") || strings.HasPrefix(url, "postgresql://") {
		idx := strings.LastIndex(url, "/")
		if idx < 0 {
			return url
		}

		rest := url[idx+1:]

		qIdx := strings.Index(rest, "?")
		if qIdx >= 0 {
			return url[:idx+1] + newName + rest[qIdx:]
		}

		return url[:idx+1] + newName
	}

	return strings.ReplaceAll(url, "dbname="+oldName, "dbname="+newName)
}

func newConfigDirect(t interface {
	Helper()
	Fatal(args ...any)
},
) Config {
	t.Helper()

	cfg, err := newConfig()
	if err != nil {
		t.Fatal("load config:", err)
	}

	return cfg
}
