// Package role_repo contains database-related CRUD functionality.
package role_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/role_create.sql
	roleCreateSql string
	//go:embed query/role_update.sql
	roleUpdateSql string
	//go:embed query/role_delete.sql
	roleDeleteSql string
	//go:embed query/role_query.sql
	roleQuerySql string
	//go:embed query/role_query_by_id.sql
	roleQueryByIdSql string
	//go:embed query/role_count.sql
	roleCountSql string
)

// Store manages the set of APIs for userDB database access.
type Store struct {
	log logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log logger.Logger, db *sqlx.DB) role.Storer {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (role.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new roleDB into the database.
func (s *Store) Create(ctx context.Context, rl role.Role) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, roleCreateSql, toRoleDB(rl))
	if err != nil {
		return fmt.Errorf("error role create in db: %w", err)
	}

	return nil
}

// Update replaces a roleDB document in the database.
func (s *Store) Update(ctx context.Context, rl role.Role) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, roleUpdateSql, toRoleDB(rl))
	if err != nil {
		return fmt.Errorf("error role update in db: %w", err)
	}

	return nil
}

// Delete removes a roleDB from the database.
func (s *Store) Delete(ctx context.Context, rl role.Role) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, roleDeleteSql, toRoleDB(rl))
	if err != nil {
		return fmt.Errorf("error delete role in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified userDB from the database.
func (s *Store) QueryByID(ctx context.Context, roleID uuid.UUID) (role.Role, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: roleID.String(),
	}

	var dbRole roleDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, roleQueryByIdSql, data, &dbRole)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return role.Role{}, fmt.Errorf("db: %w", role.ErrNotFound)
		}

		return role.Role{}, fmt.Errorf("db: %w", err)
	}

	return toRoleDomain(dbRole)
}

// Query retrieves a list of existing roles from the database.
func (s *Store) Query(
	ctx context.Context,
	filter role.QueryFilter,
	orderBy order.By,
	page web.Page,
) ([]role.Role, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	buf := bytes.NewBufferString(roleQuerySql)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbRoles []roleDB
	if err := pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbRoles); err != nil {
		return nil, fmt.Errorf("error query role in db: %w", err)
	}

	return toRolesDomain(dbRoles)
}

// Count returns the total number of users in the DB.
func (s *Store) Count(ctx context.Context, filter role.QueryFilter) (int, error) {
	data := map[string]any{}
	buf := bytes.NewBufferString(roleCountSql)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count)
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}
