package user_roles_repo

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/Housiadas/cerberus/internal/core/domain/user_roles"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/page"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

// queries
var (
	//go:embed query/user_roles_create.sql
	userRolesCreateSql string
	//go:embed query/user_roles_delete.sql
	userRolesDeleteSql string
	//go:embed query/user_roles_query.sql
	userRolesQuerySql string
	//go:embed query/user_roles_count.sql
	userRolesCountSql string
)

type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) user_roles.Storer {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new roleDB into the database.
func (s *Store) Create(ctx context.Context, rl user_roles.UserRole) error {
	if err := pgsql.NamedExecContext(ctx, s.log, s.db, userRolesCreateSql, to(rl)); err != nil {
		return fmt.Errorf("error role create in db: %w", err)
	}

	return nil
}

// Delete removes a roleDB from the database.
func (s *Store) Delete(ctx context.Context, rl user_roles.UserRole) error {
	if err := pgsql.NamedExecContext(ctx, s.log, s.db, userRolesDeleteSql, toRoleDB(rl)); err != nil {
		return fmt.Errorf("error delete role in db: %w", err)
	}

	return nil
}

// Query retrieves a list of existing roles from the database.
func (s *Store) Query(
	ctx context.Context,
	filter user_roles.QueryFilter,
	orderBy order.By,
	page page.Page,
) ([]user_roles.UserRole, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	buf := bytes.NewBufferString(userRolesQuerySql)
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
func (s *Store) Count(ctx context.Context, filter user_roles.QueryFilter) (int, error) {
	data := map[string]any{}
	buf := bytes.NewBufferString(userRolesCountSql)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := pgsql.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}
