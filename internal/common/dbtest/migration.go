package dbtest

import (
	"database/sql"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func migration(cfg Config, dbURL string) error {
	db, err := sql.Open("postgres", dbURL+"sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
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
