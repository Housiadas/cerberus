package dbtest

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func migration(cfg Config, DBName string) error {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", cfg.DBUser, cfg.DBPassword, cfg.DBPort, DBName))

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		getMigrationsDir(),
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}
	// or m.Step(2) if you want to explicitly set the number of .migrations to run
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
