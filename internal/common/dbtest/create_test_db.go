package dbtest

import (
	"context"
	"math/rand"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/Housiadas/cerberus/pkg/pgsql"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

type TestDB struct {
	DB   *sqlx.DB
	Name string
}

func CreateTestDB(
	t *testing.T,
	cfg Config,
	host string,
	dbManagement *sqlx.DB,
) TestDB {
	b := make([]byte, 4)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	dbName := string(b)

	t.Logf("Create Database: %s\n", dbName)
	if _, err := dbManagement.ExecContext(context.Background(), "CREATE DATABASE "+dbName); err != nil {
		t.Fatalf("[TEST]: creating database %s: %v", dbName, err)
	}

	db, err := pgsql.Open(pgsql.Config{
		User:       cfg.DBUser,
		Password:   cfg.DBPassword,
		Host:       host,
		Name:       dbName,
		DisableTLS: true,
	})
	if err != nil {
		t.Fatalf("[TEST]: Opening database connection: %v", err)
	}

	return TestDB{
		DB:   db,
		Name: dbName,
	}
}
