// Package permission_repo contains database-related CRUD functionality.
package permission_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	permission2 "github.com/Housiadas/cerberus/internal/core/permission"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/permission_create.sql
	permissionCreateSQL string
	//go:embed query/permission_update.sql
	permissionUpdateSQL string
	//go:embed query/permission_delete.sql
	permissionDeleteSQL string
	//go:embed query/permission_query.sql
	permissionQuerySQL string
	//go:embed query/permission_query_by_id.sql
	permissionQueryByIDSQL string
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (permission2.Storer, error) {
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
func (s *Store) Create(ctx context.Context, perm permission2.Permission) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, permissionCreateSQL, toPermissionDB(perm))
	if err != nil {
		return fmt.Errorf("error permission create in db: %w", err)
	}

	return nil
}

// Update replaces a permissionDB document in the database.
func (s *Store) Update(ctx context.Context, rl permission2.Permission) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, permissionUpdateSQL, toPermissionDB(rl))
	if err != nil {
		return fmt.Errorf("error permission update in db: %w", err)
	}

	return nil
}

// Delete removes a permissionDB from the database.
func (s *Store) Delete(ctx context.Context, rl permission2.Permission) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, permissionDeleteSQL, toPermissionDB(rl))
	if err != nil {
		return fmt.Errorf("error delete permission in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified userDB from the database.
func (s *Store) QueryByID(
	ctx context.Context,
	permissionID uuid.UUID,
) (permission2.Permission, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: permissionID.String(),
	}

	var dbPermission permissionDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.dbPool, permissionQueryByIDSQL, data, &dbPermission)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return permission2.Permission{}, fmt.Errorf("db: %w", permission2.ErrNotFound)
		}

		return permission2.Permission{}, fmt.Errorf("db: %w", err)
	}

	return toPermissionDomain(dbPermission)
}

// Query retrieves a list of existing permissions from the database.
func (s *Store) Query(
	ctx context.Context,
	filter permission2.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]permission2.Permission, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(permissionQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, fmt.Errorf("permission order issue: %w", err)
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbPermissions []permissionDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.dbPool, buf.String(), data, &dbPermissions)
	if err != nil {
		return nil, fmt.Errorf("error query permission in db: %w", err)
	}

	return toPermissionsDomain(dbPermissions)
}
