// Package permission_repo contains database-related CRUD functionality.
package permission_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/permission_create.sql
	permissionCreateSql string
	//go:embed query/permission_update.sql
	permissionUpdateSql string
	//go:embed query/permission_delete.sql
	permissionDeleteSql string
	//go:embed query/permission_query.sql
	permissionQuerySql string
	//go:embed query/permission_query_by_id.sql
	permissionQueryByIdSql string
	//go:embed query/permission_count.sql
	permissionCountSql string
)

// Store manages the set of APIs for userDB database access.
type Store struct {
	log    logger.Logger
	dbPool sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log logger.Logger, dbPool *sqlx.DB) permission.Storer {
	return &Store{
		log:    log,
		dbPool: dbPool,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (permission.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("permission init transaction error: %w", err)
	}

	store := Store{
		log:    s.log,
		dbPool: ec,
	}

	return &store, nil
}

// Create inserts a new permissionDB into the database.
func (s *Store) Create(ctx context.Context, perm permission.Permission) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, permissionCreateSql, toPermissionDB(perm))
	if err != nil {
		return fmt.Errorf("error permission create in db: %w", err)
	}

	return nil
}

// Update replaces a permissionDB document in the database.
func (s *Store) Update(ctx context.Context, rl permission.Permission) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, permissionUpdateSql, toPermissionDB(rl))
	if err != nil {
		return fmt.Errorf("error permission update in db: %w", err)
	}

	return nil
}

// Delete removes a permissionDB from the database.
func (s *Store) Delete(ctx context.Context, rl permission.Permission) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, permissionDeleteSql, toPermissionDB(rl))
	if err != nil {
		return fmt.Errorf("error delete permission in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified userDB from the database.
func (s *Store) QueryByID(
	ctx context.Context,
	permissionID uuid.UUID,
) (permission.Permission, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: permissionID.String(),
	}

	var dbPermission permissionDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.dbPool, permissionQueryByIdSql, data, &dbPermission)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return permission.Permission{}, fmt.Errorf("db: %w", permission.ErrNotFound)
		}

		return permission.Permission{}, fmt.Errorf("db: %w", err)
	}

	return toPermissionDomain(dbPermission)
}

// Query retrieves a list of existing permissions from the database.
func (s *Store) Query(
	ctx context.Context,
	filter permission.QueryFilter,
	orderBy order.By,
	page web.Page,
) ([]permission.Permission, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	buf := bytes.NewBufferString(permissionQuerySql)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, fmt.Errorf("permission order issue: %w", err)
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbPermissions []permissionDB
	if err := pgsql.NamedQuerySlice(
		ctx,
		s.log,
		s.dbPool,
		buf.String(),
		data,
		&dbPermissions,
	); err != nil {
		return nil, fmt.Errorf("error query permission in db: %w", err)
	}

	return toPermissionsDomain(dbPermissions)
}

// Count returns the total number of users in the DB.
func (s *Store) Count(ctx context.Context, filter permission.QueryFilter) (int, error) {
	data := map[string]any{}
	buf := bytes.NewBufferString(permissionCountSql)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}

	err := pgsql.NamedQueryStruct(ctx, s.log, s.dbPool, buf.String(), data, &count)
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}
