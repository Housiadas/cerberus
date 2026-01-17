// Package user_roles_permissions_repo contains DB access for the user_roles_permissions view.
package user_roles_permissions_repo

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	urp "github.com/Housiadas/cerberus/internal/core/domain/user_roles_permissions"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/user_roles_permissions_query.sql
	userRolesPermissionsQuerySQL string
	//go:embed query/user_roles_permissions_count.sql
	userRolesPermissionsCountSQL string
	//go:embed query/user_roles_permissions_count.sql
	userHasPermissionSQL string
)

// Store manages the set of APIs for DB access to the view.
type Store struct {
	log logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Query retrieves rows from the view with paging.
func (s *Store) Query(
	ctx context.Context,
	filter urp.QueryFilter,
	ob order.By,
	p web.Page,
) ([]urp.UserRolesPermissions, error) {
	data := map[string]any{
		"offset":        (p.Number() - 1) * p.RowsPerPage(),
		"rows_per_page": p.RowsPerPage(),
	}

	buf := bytes.NewBufferString(userRolesPermissionsQuerySQL)
	applyFilter(filter, data, buf)

	obc, err := orderByClause(ob)
	if err != nil {
		return nil, err
	}

	buf.WriteString(obc)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbRows []rowDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbRows)
	if err != nil {
		return nil, fmt.Errorf("db query user_roles_permissions: %w", err)
	}

	return toDomains(dbRows)
}

// Count returns the total number of rows that match the filter.
func (s *Store) Count(ctx context.Context, filter urp.QueryFilter) (int, error) {
	data := map[string]any{}
	buf := bytes.NewBufferString(userRolesPermissionsCountSQL)
	applyFilter(filter, data, buf)

	var cnt struct {
		Count int `db:"count"`
	}

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &cnt)
	if err != nil {
		return 0, fmt.Errorf("db count user_roles_permissions: %w", err)
	}

	return cnt.Count, nil
}

// HasPermission returns true if the user has the specified permission.
func (s *Store) HasPermission(
	ctx context.Context,
	userID uuid.UUID,
	permissionName string,
) (bool, error) {
	data := map[string]any{
		"user_id":         userID,
		"permission_name": permissionName,
	}

	pName, err := name.Parse(permissionName)
	if err != nil {
		return false, fmt.Errorf("parse name: %w", err)
	}

	filter := urp.QueryFilter{
		UserID:         &userID,
		PermissionName: &pName,
	}

	buf := bytes.NewBufferString(userHasPermissionSQL)
	applyFilter(filter, data, buf)

	var cnt struct {
		Count int `db:"count"`
	}

	err = pgsql.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &cnt)
	if err != nil {
		return false, fmt.Errorf("db count has permissions: %w", err)
	}

	return cnt.Count >= 1, nil
}
