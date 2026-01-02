// Package dbtest contains supporting code for running tests that hit the DB.
package dbtest

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/Housiadas/cerberus/pkg/logger"
)

// Database owns the state for running and shutting down tests.
type Database struct {
	DB   *sqlx.DB
	Log  *logger.Logger
	Core Service
}

// New creates a new test database inside the database that was started
// to handle testing. The database is migrated to the current version, and
// a connection pool is provided with internal core packages.
func New(t *testing.T, testName string) *Database {
	// load app local config
	cfg := newConfig(t)

	// Start the postgres container and run any migrations on it
	ctx := context.Background()

	ctr, err := postgres.Run(
		ctx,
		cfg.PostgresImage,
		postgres.WithDatabase(cfg.DBName),
		postgres.WithUsername(cfg.DBUser),
		postgres.WithPassword(cfg.DBPassword),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// database url
	dbURL, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	// set up migrations
	err = migration(cfg, dbURL)
	require.NoError(t, err)

	// Create a snapshot of the database to restore later
	err = ctr.Snapshot(ctx, postgres.WithSnapshotName("test-snapshot"))
	require.NoError(t, err)

	// Open DB
	dbTest, err := sqlx.Open("pgx", dbURL)
	require.NoError(t, err)

	// inject logger
	var buf bytes.Buffer
	traceIDfn := func(context.Context) string { return "" }
	requestIDfn := func(context.Context) string { return "" }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDfn, requestIDfn)

	// -------------------------------------------------------------------------
	// should be invoked when the caller is done with the database
	t.Cleanup(func() {
		t.Helper()

		// Close the DB
		dbTest.Close()

		// Connect to a postgres database to terminate connections
		postgresURL := strings.Replace(dbURL, cfg.DBName, "postgres", 1)
		adminDB, err := sqlx.Open("pgx", postgresURL)
		require.NoError(t, err)
		defer adminDB.Close()

		_, err = adminDB.Exec(fmt.Sprintf(`
      SELECT pg_terminate_backend(pg_stat_activity.pid)
      FROM pg_stat_activity
      WHERE pg_stat_activity.datname = '%s'
      AND pid <> pg_backend_pid()
   `, cfg.DBName))
		require.NoError(t, err)

		// Reset the DB to its snapshot state.
		err = ctr.Restore(ctx)
		require.NoError(t, err)

		t.Logf("******************** LOGS (%s) ********************\n\n", testName)
		t.Log(buf.String())
		t.Logf("******************** LOGS (%s) ********************\n", testName)
	})

	return &Database{
		DB:   dbTest,
		Log:  log,
		Core: newCore(log, dbTest),
	}
}
