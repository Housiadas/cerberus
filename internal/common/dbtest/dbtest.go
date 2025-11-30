// Package dbtest contains supporting code for running tests that hit the DB.
package dbtest

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	ctxPck "github.com/Housiadas/cerberus/internal/common/context"
	"github.com/Housiadas/cerberus/pkg/docker"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/otel"
	"github.com/Housiadas/cerberus/pkg/pgsql"
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
	// -------------------------------------------------------------------------
	// load app local config
	cfg := newConfig(t)

	// -------------------------------------------------------------------------
	// start container
	dockerArgs := []string{
		"-e", fmt.Sprintf("POSTGRES_DB=%s", cfg.DBName),
		"-e", fmt.Sprintf("POSTGRES_USER=%s", cfg.DBUser),
		"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", cfg.DBPassword),
	}
	appArgs := []string{"-c", "log_statement=all"}

	c, err := docker.StartContainer(cfg.PostgresImage, cfg.PostgresContainerName, cfg.DBPort, dockerArgs, appArgs)
	if err != nil {
		t.Fatalf("[TEST]: Starting database: %v", err)
	}

	t.Logf("Name    : %s\n", c.Name)
	t.Logf("Host: %s\n", c.HostPort)

	// -------------------------------------------------------------------------
	// open management db
	dbManagement, err := pgsql.Open(pgsql.Config{
		User:       cfg.DBUser,
		Password:   cfg.DBPassword,
		Host:       c.HostPort,
		Name:       cfg.DBName,
		DisableTLS: true,
	})
	if err != nil {
		t.Fatalf("[TEST]: Opening database connection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pgsql.StatusCheck(ctx, dbManagement); err != nil {
		t.Fatalf("[TEST]: status check database: %v", err)
	}

	// -------------------------------------------------------------------------
	// open test db
	testDB := CreateTestDB(t, cfg, c.HostPort, dbManagement)

	// -------------------------------------------------------------------------
	// set up migrations
	t.Logf("[TEST]: migrate Database UP %s\n", testDB.Name)
	err = migration(cfg, testDB.Name)
	if err != nil {
		t.Fatalf("[TEST]: Migrating error: %s", err)
	}

	// -------------------------------------------------------------------------
	// inject logger
	var buf bytes.Buffer
	traceIDfn := func(context.Context) string { return otel.GetTraceID(ctx) }
	requestIDfn := func(context.Context) string { return ctxPck.GetRequestID(ctx) }
	log := logger.New(&buf, logger.LevelInfo, "TEST", traceIDfn, requestIDfn)

	// -------------------------------------------------------------------------
	// should be invoked when the caller is done with the database.
	t.Cleanup(func() {
		t.Helper()

		t.Logf("[TEST]: Drop Database: %s\n", testDB.Name)
		if _, err := dbManagement.ExecContext(context.Background(), "DROP DATABASE "+testDB.Name+" WITH (force)"); err != nil {
			t.Fatalf("[TEST]: dropping database %s: %v", testDB.Name, err)
		}

		testDB.DB.Close()
		dbManagement.Close()

		t.Logf("******************** LOGS (%s) ********************\n\n", testName)
		t.Log(buf.String())
		t.Logf("******************** LOGS (%s) ********************\n", testName)
	})

	return &Database{
		DB:   testDB.DB,
		Log:  log,
		Core: newCore(log, testDB.DB),
	}
}
