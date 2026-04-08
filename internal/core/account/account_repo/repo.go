// Package account_repo contains database related CRUD functionality.
package account_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/account_create.sql
	accountCreateSQL string
	//go:embed query/account_update.sql
	accountUpdateSQL string
	//go:embed query/account_delete.sql
	accountDeleteSQL string
	//go:embed query/account_query.sql
	accountQuerySQL string
	//go:embed query/account_query_by_id.sql
	accountQueryByIDSQL string
)

// Store manages the set of APIs for account database access.
type Store struct {
	log logger.Logger
	db  *sqlx.DB
}

// NewStore constructs the api for data access.
func NewStore(
	log logger.Logger,
	db *sqlx.DB,
) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new account into the database.
func (s *Store) Create(ctx context.Context, acc account.Account) error {
	err := pgsql.NamedExecContext(
		ctx,
		s.log,
		pgsql.Conn(ctx, s.db),
		accountCreateSQL,
		toAccountDB(acc),
	)
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Update replaces an account document in the database.
func (s *Store) Update(ctx context.Context, acc account.Account) error {
	err := pgsql.NamedExecContext(
		ctx,
		s.log,
		pgsql.Conn(ctx, s.db),
		accountUpdateSQL,
		toAccountDB(acc),
	)
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Delete removes an account from the database.
func (s *Store) Delete(ctx context.Context, acc account.Account) error {
	err := pgsql.NamedExecContext(
		ctx,
		s.log,
		pgsql.Conn(ctx, s.db),
		accountDeleteSQL,
		toAccountDB(acc),
	)
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Query retrieves a list of existing accounts from the database.
func (s *Store) Query(
	ctx context.Context,
	filter account.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]account.Account, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(accountQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbAccs []accountDB

	err = pgsql.NamedQuerySlice(ctx, s.log, pgsql.Conn(ctx, s.db), buf.String(), data, &dbAccs)
	if err != nil {
		return nil, fmt.Errorf("named_query_slice: %w", err)
	}

	return toAccountsDomain(dbAccs), nil
}

// QueryByID gets the specified account from the database.
func (s *Store) QueryByID(ctx context.Context, accountID uuid.UUID) (account.Account, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: accountID.String(),
	}

	var dbAcc accountDB

	err := pgsql.NamedQueryStruct(
		ctx,
		s.log,
		pgsql.Conn(ctx, s.db),
		accountQueryByIDSQL,
		data,
		&dbAcc,
	)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return account.Account{}, fmt.Errorf("db: %w", account.ErrNotFound)
		}

		return account.Account{}, fmt.Errorf("db: %w", err)
	}

	return toAccountDomain(dbAcc), nil
}
