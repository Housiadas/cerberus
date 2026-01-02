package dbtest

import (
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

func migration(cfg Config, dbTest *sqlx.DB) error {
	driver, err := postgres.WithInstance(dbTest.DB, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		getMigrationsDir(),
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil {
		return err
	}

	return nil
}

func getMigrationsDir() string {
	_, file, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(file)
	migrationsPath := filepath.Join(basepath, "../../../.migrations")
	return "file://" + migrationsPath
}
