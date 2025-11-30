// Package role_repo contains database-related CRUD functionality.
package role_repo

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

// queries
var (
	//go:embed query/role_create.sql
	roleCreateSql string
	//go:embed query/role_update.sql
	roleUpdateSql string
	//go:embed query/role_delete.sql
	roleDeleteSql string
	//go:embed query/role_query.sql
	roleQuerySql string
)

// Store manages the set of APIs for userDB database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
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
	if err := pgsql.NamedExecContext(ctx, s.log, s.db, roleCreateSql, toRoleDB(rl)); err != nil {
		return fmt.Errorf("error role create in db: %w", err)
	}

	return nil
}

// Update replaces a roleDB document in the database.
func (s *Store) Update(ctx context.Context, rl role.Role) error {
	if err := pgsql.NamedExecContext(ctx, s.log, s.db, roleUpdateSql, toRoleDB(rl)); err != nil {
		return fmt.Errorf("error role update in db: %w", err)
	}

	return nil
}

// Delete removes a roleDB from the database.
func (s *Store) Delete(ctx context.Context, rl role.Role) error {
	if err := pgsql.NamedExecContext(ctx, s.log, s.db, roleDeleteSql, toRoleDB(rl)); err != nil {
		return fmt.Errorf("error delete role in db: %w", err)
	}

	return nil
}

// Query retrieves a list of existing roles from the database.
func (s *Store) Query(
	ctx context.Context,
	filter role.QueryFilter,
	orderBy order.By,
	page page.Page,
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
