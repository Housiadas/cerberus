// Package user_repo contains database related CRUD functionality.
package user_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/mail"

	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/user_create.sql
	userCreateSQL string
	//go:embed query/user_update.sql
	userUpdateSQL string
	//go:embed query/user_delete.sql
	userDeleteSQL string
	//go:embed query/user_query.sql
	userQuerySQL string
	//go:embed query/user_query_by_id.sql
	userQueryByIDSQL string
	//go:embed query/user_query_by_email.sql
	userQueryByEmailSQL string
)

// Store manages the set of APIs for userDB database access.
type Store struct {
	log    logger.Logger
	dbPool sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log logger.Logger, dbPool *sqlx.DB) *Store {
	return &Store{
		log:    log,
		dbPool: dbPool,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (user.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("user transaction init error: %w", err)
	}

	store := Store{
		log:    s.log,
		dbPool: ec,
	}

	return &store, nil
}

// Create inserts a new userDB into the database.
func (s *Store) Create(ctx context.Context, usr user.User) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, userCreateSQL, toUserDB(usr))
	if err != nil {
		if errors.Is(err, pgsql.ErrDBDuplicatedEntry) {
			return fmt.Errorf("named_exec_context: %w", user.ErrUniqueEmail)
		}

		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Update replaces a userDB document in the database.
func (s *Store) Update(ctx context.Context, usr user.User) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, userUpdateSQL, toUserDB(usr))
	if err != nil {
		if errors.Is(err, pgsql.ErrDBDuplicatedEntry) {
			return user.ErrUniqueEmail
		}

		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Delete removes a userDB from the database.
func (s *Store) Delete(ctx context.Context, usr user.User) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, userDeleteSQL, toUserDB(usr))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (s *Store) Query(
	ctx context.Context,
	filter user.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]user.User, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(userQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbUsrs []userDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.dbPool, buf.String(), data, &dbUsrs)
	if err != nil {
		return nil, fmt.Errorf("named_query_slice: %w", err)
	}

	return toUsersDomain(dbUsrs)
}

// QueryByID gets the specified userDB from the database.
func (s *Store) QueryByID(ctx context.Context, userID uuid.UUID) (user.User, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: userID.String(),
	}

	var dbUsr userDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.dbPool, userQueryByIDSQL, data, &dbUsr)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return user.User{}, fmt.Errorf("db: %w", user.ErrNotFound)
		}

		return user.User{}, fmt.Errorf("db: %w", err)
	}

	return toUserDomain(dbUsr)
}

// QueryByEmail gets the specified userDB from the database by email.
func (s *Store) QueryByEmail(ctx context.Context, email mail.Address) (user.User, error) {
	data := struct {
		Email string `db:"email"`
	}{
		Email: email.Address,
	}

	var dbUsr userDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.dbPool, userQueryByEmailSQL, data, &dbUsr)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return user.User{}, fmt.Errorf("db: %w", user.ErrNotFound)
		}

		return user.User{}, fmt.Errorf("db: %w", err)
	}

	return toUserDomain(dbUsr)
}
