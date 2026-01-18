// Package pgsql provides support for access the database.
package pgsql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/otel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
)

// lib/pq errorCodeNames
// https://github.com/lib/pq/blob/master/error.go#L178
const (
	uniqueViolation = "23505"
	undefinedTable  = "42P01"
)

// Set of error variables for CRUD operations.
var (
	ErrDBNotFound        = sql.ErrNoRows
	ErrDBDuplicatedEntry = errors.New("duplicated entry")
	ErrUndefinedTable    = errors.New("undefined table")
)

// Config is the required properties to use the database.
type Config struct {
	User         string
	Password     string
	Host         string
	Name         string
	Schema       string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg Config) (*sqlx.DB, error) {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	if cfg.Schema != "" {
		q.Set("search_path", cfg.Schema)
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	dbPool, err := sqlx.Open("pgx", u.String())
	if err != nil {
		return nil, fmt.Errorf("sqlx driver open error: %w", err)
	}

	dbPool.SetMaxIdleConns(cfg.MaxIdleConns)
	dbPool.SetMaxOpenConns(cfg.MaxOpenConns)

	return dbPool, nil
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, dbPool *sqlx.DB) error {
	// If the user doesn't give us a deadline set 1 second.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
	}

	for attempts := 1; ; attempts++ {
		err := dbPool.PingContext(ctx)
		if err == nil {
			break
		}

		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)

		if ctx.Err() != nil {
			//nolint:wrapcheck
			return ctx.Err()
		}
	}

	if ctx.Err() != nil {
		return fmt.Errorf("context error: %w", ctx.Err())
	}

	// Run a simple query to determine connectivity.
	// Running this query forces a round trip through the database.
	const q = `SELECT TRUE`

	var tmp bool

	err := dbPool.QueryRowContext(ctx, q).Scan(&tmp)

	return fmt.Errorf("query row context error: %w", err)
}

// ExecContext is a helper function to execute a CUD operation with
// logging and tracing.
func ExecContext(ctx context.Context, log logger.Logger, db sqlx.ExtContext, query string) error {
	return NamedExecContext(ctx, log, db, query, struct{}{})
}

// NamedExecContext is a helper function to execute a CRUD operation with
// logging and tracing where field replacement is necessary.
func NamedExecContext(
	ctx context.Context,
	log logger.Logger,
	db sqlx.ExtContext,
	query string,
	data any,
) error {
	q := queryString(query, data)

	var err error

	defer func() {
		if err != nil {
			switch data.(type) {
			case struct{}:
				log.Infoc(ctx, 6, "database.NamedExecContext", "query", q, "ERROR", err)
			default:
				log.Infoc(ctx, 5, "database.NamedExecContext", "query", q, "ERROR", err)
			}
		}
	}()

	ctx, span := otel.AddSpan(ctx, "internal.api.pgsql.exec", attribute.String("query", q))
	defer span.End()

	_, err = sqlx.NamedExecContext(ctx, db, query, data)
	if err != nil {
		var pqerr *pgconn.PgError
		if errors.As(err, &pqerr) {
			switch pqerr.Code {
			case undefinedTable:
				return ErrUndefinedTable
			case uniqueViolation:
				return ErrDBDuplicatedEntry
			}
		}

		return fmt.Errorf("pg error: %w", err)
	}

	return nil
}

// QuerySlice is a helper function for executing queries that return a
// collection of data to be unmarshalled into a slice.
func QuerySlice[T any](
	ctx context.Context,
	log *logger.Service,
	db sqlx.ExtContext,
	query string,
	dest *[]T,
) error {
	return namedQuerySlice(ctx, log, db, query, struct{}{}, dest, false)
}

// NamedQuerySlice is a helper function for executing queries that return a
// collection of data to be unmarshalled into a slice where field replacement is
// necessary.
func NamedQuerySlice[T any](
	ctx context.Context,
	log logger.Logger,
	db sqlx.ExtContext,
	query string,
	data any,
	dest *[]T,
) error {
	return namedQuerySlice(ctx, log, db, query, data, dest, false)
}

// NamedQuerySliceUsingIn is a helper function for executing queries that return
// a collection of data to be unmarshalled into a slice where field replacement
// is necessary. Use this if the query has an IN clause.
func NamedQuerySliceUsingIn[T any](
	ctx context.Context,
	log *logger.Service,
	db sqlx.ExtContext,
	query string,
	data any,
	dest *[]T,
) error {
	return namedQuerySlice(ctx, log, db, query, data, dest, true)
}

//nolint:cyclop
func namedQuerySlice[T any](
	ctx context.Context,
	log logger.Logger,
	db sqlx.ExtContext,
	query string,
	data any,
	dest *[]T,
	withIn bool,
) error {
	var err error

	q := queryString(query, data)

	defer func() {
		if err != nil {
			log.Infoc(ctx, 6, "database.NamedQuerySlice", "query", q, "ERROR", err)
		}
	}()

	ctx, span := otel.AddSpan(ctx, "internal.api.pgsql.queryslice", attribute.String("query", q))
	defer span.End()

	var rows *sqlx.Rows

	switch withIn {
	case true:
		rows, err = func() (*sqlx.Rows, error) {
			named, args, err := sqlx.Named(query, data)
			if err != nil {
				return nil, fmt.Errorf("sqlx named error: %w", err)
			}

			query, args, err := sqlx.In(named, args...)
			if err != nil {
				return nil, fmt.Errorf("sqlx in: %w", err)
			}

			query = db.Rebind(query)

			return db.QueryxContext(ctx, query, args...)
		}()

	default:
		rows, err = sqlx.NamedQueryContext(ctx, db, query, data)
	}

	if err != nil {
		var pqerr *pgconn.PgError
		if errors.As(err, &pqerr) && pqerr.Code == undefinedTable {
			return ErrUndefinedTable
		}

		return fmt.Errorf("pg error: %w", err)
	}

	defer rows.Close()

	var slice []T

	for rows.Next() {
		v := new(T)

		err := rows.StructScan(v)
		if err != nil {
			return fmt.Errorf("struct scan error: %w", err)
		}

		slice = append(slice, *v)
	}

	*dest = slice

	return nil
}

// QueryStruct is a helper function for executing queries that return a
// single value to be unmarshalled into a struct type where field replacement is necessary.
func QueryStruct(
	ctx context.Context,
	log *logger.Service,
	db sqlx.ExtContext,
	query string,
	dest any,
) error {
	return namedQueryStruct(ctx, log, db, query, struct{}{}, dest, false)
}

// NamedQueryStruct is a helper function for executing queries that return a
// single value to be unmarshalled into a struct type where field replacement is necessary.
func NamedQueryStruct(
	ctx context.Context,
	log logger.Logger,
	db sqlx.ExtContext,
	query string,
	data any,
	dest any,
) error {
	return namedQueryStruct(ctx, log, db, query, data, dest, false)
}

// NamedQueryStructUsingIn is a helper function for executing queries that return
// a single value to be unmarshalled into a struct type where field replacement
// is necessary. Use this if the query has an IN clause.
func NamedQueryStructUsingIn(
	ctx context.Context,
	log logger.Logger,
	db sqlx.ExtContext,
	query string,
	data any,
	dest any,
) error {
	return namedQueryStruct(ctx, log, db, query, data, dest, true)
}

//nolint:cyclop
func namedQueryStruct(
	ctx context.Context,
	log logger.Logger,
	db sqlx.ExtContext,
	query string,
	data any,
	dest any,
	withIn bool,
) error {
	var err error

	q := queryString(query, data)

	defer func() {
		if err != nil {
			log.Infoc(ctx, 6, "database.NamedQuerySlice", "query", q, "ERROR", err)
		}
	}()

	ctx, span := otel.AddSpan(ctx, "internal.api.pgsql.query", attribute.String("query", q))
	defer span.End()

	var rows *sqlx.Rows

	switch withIn {
	case true:
		rows, err = func() (*sqlx.Rows, error) {
			named, args, err := sqlx.Named(query, data)
			if err != nil {
				return nil, fmt.Errorf("sqlx named error: %w", err)
			}

			query, args, err := sqlx.In(named, args...)
			if err != nil {
				return nil, fmt.Errorf("sqlx in error: %w", err)
			}

			query = db.Rebind(query)

			return db.QueryxContext(ctx, query, args...)
		}()

	default:
		rows, err = sqlx.NamedQueryContext(ctx, db, query, data)
	}

	if err != nil {
		var pqerr *pgconn.PgError
		if errors.As(err, &pqerr) && pqerr.Code == undefinedTable {
			return ErrUndefinedTable
		}

		return fmt.Errorf("pg driver error: %w", err)
	}

	defer rows.Close()

	if !rows.Next() {
		return ErrDBNotFound
	}

	err = rows.StructScan(dest)
	if err != nil {
		return fmt.Errorf("struct scan error: %w", err)
	}

	return nil
}

// queryString provides a pretty print version of the query and parameters.
func queryString(query string, args any) string {
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return err.Error()
	}

	for _, param := range params {
		var value string

		switch paramType := param.(type) {
		case string:
			value = fmt.Sprintf("'%s'", paramType)
		case []byte:
			value = fmt.Sprintf("'%s'", string(paramType))
		default:
			value = fmt.Sprintf("%v", paramType)
		}

		query = strings.Replace(query, "?", value, 1)
	}

	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")

	return strings.Trim(query, " ")
}
