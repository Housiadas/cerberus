package dbtest

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // This is for golang-migrate
	_ "github.com/jackc/pgx/v5/stdlib"                   // This is for golang-migrate sqlx driver
)

func migration(dbURL string) error {
	db, err := sql.Open("pgx", dbURL+"sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		getMigrationsDir(),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

func getMigrationsDir() string {
	_, file, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(file)
	migrationsPath := filepath.Join(basepath, "../../../.migrations")

	return "file://" + migrationsPath
}
