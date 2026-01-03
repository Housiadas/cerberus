// Package dbtest contains supporting code for running tests that hit the DB.
package dbtest

import (
	"bytes"
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// Database owns the state for running and shutting down tests.
type Database struct {
	DB *sqlx.DB
}

// New creates a new test database inside the database that was started
// to handle testing. The database is migrated to the current version, and
// a connection pool is provided with internal core packages.
func New(t *testing.T, testName string) *sqlx.DB {
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

	// Open DB
	dbTest, err := sqlx.Open("pgx", dbURL)
	require.NoError(t, err)

	// -------------------------------------------------------------------------
	// Will be invoked when the caller is done with the database
	var buf bytes.Buffer
	t.Cleanup(func() {
		t.Helper()

		// Close the DB
		dbTest.Close()

		t.Logf("******************** LOGS (%s) ********************\n\n", testName)
		t.Log(buf.String())
		t.Logf("******************** LOGS (%s) ********************\n", testName)
	})

	return dbTest
}
