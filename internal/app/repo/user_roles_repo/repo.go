package user_roles_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/Housiadas/cerberus/internal/core/domain/user_roles"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
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
	//go:embed query/user_roles_view_by_user_id.sql
	userRolesByUserIdSql string
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
	if err := pgsql.NamedExecContext(ctx, s.log, s.db, userRolesCreateSql, toUserRoleDB(rl)); err != nil {
		return fmt.Errorf("error role create in db: %w", err)
	}

	return nil
}

// Delete removes a roleDB from the database.
func (s *Store) Delete(ctx context.Context, rl user_roles.UserRole) error {
	if err := pgsql.NamedExecContext(ctx, s.log, s.db, userRolesDeleteSql, toUserRoleDB(rl)); err != nil {
		return fmt.Errorf("error delete role in db: %w", err)
	}

	return nil
}

// Query retrieves a list of existing roles from the database.
func (s *Store) Query(
	ctx context.Context,
	filter user_roles.QueryFilter,
	orderBy order.By,
	page web.Page,
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

	var dbRoles []userRolesDB
	if err := pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbRoles); err != nil {
		return nil, fmt.Errorf("error query role in db: %w", err)
	}

	return toDomains(dbRoles), nil
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

func (s *Store) GetUserRoleNames(ctx context.Context, userID uuid.UUID) ([]string, error) {
	data := map[string]any{
		"user_id": userID,
	}
	filter := user_roles.QueryFilter{
		UserID: &userID,
	}

	buf := bytes.NewBufferString(userRolesByUserIdSql)
	applyFilter(filter, data, buf)

	var dbUsrRolesView []userRolesViewDB
	if err := pgsql.NamedQuerySlice(ctx, s.log, s.db, userRolesByUserIdSql, data, &dbUsrRolesView); err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return nil, fmt.Errorf("db: %w", user_roles.ErrNotFound)
		}
		return nil, fmt.Errorf("db: %w", err)
	}

	return toUserRoleNames(dbUsrRolesView), nil
}
