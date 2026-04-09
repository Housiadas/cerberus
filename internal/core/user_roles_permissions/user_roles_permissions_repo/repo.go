// Package user_roles_permissions_repo contains DB access for the user_roles_permissions view.
package user_roles_permissions_repo

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/user_roles_permissions"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/user_permissions_by_user_id.sql
	userPermissionsByUserIDSQL string
)

// Store manages the set of APIs for DB access to the view.
type Store struct {
	log logger.Logger
	db  sqlx.ExtContext
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

// Query retrieves rows from the view with paging.
func (s *Store) Query(
	ctx context.Context,
	filter user_roles_permissions.QueryFilter,
	ob order.By,
	cur cursor.Cursor,
) ([]user_roles_permissions.UserRolesPermissions, error) {
	col, ok := orderByFields[ob.Field]
	if !ok {
		return nil, fmt.Errorf("%w: %s", errOrderFieldNotFound, ob.Field)
	}

	sb := pgsql.Builder.
		Select(
			"user_id", "user_name", "user_email",
			"role_id", "role_name",
			"permission_id", "permission_name",
		).
		From("vw_user_roles_permissions").
		Where(filterPredicates(filter)).
		OrderBy(col+" "+ob.Direction, "user_id "+ob.Direction).
		Limit(uint64(cur.Limit() + 1))

	if cp := cursorPredicate(cur, ob); cp != nil {
		sb = sb.Where(cp)
	}

	query, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var dbRows []rowDB

	if err := pgsql.SelectSlice(ctx, s.log, s.db, query, args, &dbRows); err != nil {
		return nil, fmt.Errorf("select slice: %w", err)
	}

	return toDomains(dbRows)
}

// QueryPermissionsByUserID returns all distinct permissions (id and name) for the given user.
func (s *Store) QueryPermissionsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]user_roles_permissions.Permission, error) {
	data := map[string]any{
		"user_id": userID,
	}

	var rows []struct {
		PermissionID   uuid.UUID `db:"permission_id"`
		PermissionName string    `db:"permission_name"`
	}

	err := pgsql.NamedQuerySlice(ctx, s.log, s.db, userPermissionsByUserIDSQL, data, &rows)
	if err != nil {
		return nil, fmt.Errorf("db query permissions by user_id: %w", err)
	}

	permissions := make([]user_roles_permissions.Permission, len(rows))
	for i, r := range rows {
		permissions[i] = user_roles_permissions.Permission{
			ID:   r.PermissionID,
			Name: r.PermissionName,
		}
	}

	return permissions, nil
}

// HasPermission returns true if the user has the specified permission.
func (s *Store) HasPermission(
	ctx context.Context,
	userID uuid.UUID,
	permissionName string,
) (bool, error) {
	pName, err := name.Parse(permissionName)
	if err != nil {
		return false, fmt.Errorf("parse name: %w", err)
	}

	filter := user_roles_permissions.QueryFilter{
		UserID:         &userID,
		PermissionName: &pName,
	}

	query, args, err := pgsql.Builder.
		Select("COUNT(1) AS count").
		From("vw_user_roles_permissions").
		Where(filterPredicates(filter)).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("build query: %w", err)
	}

	var rows []struct {
		Count int `db:"count"`
	}

	err = pgsql.SelectSlice(ctx, s.log, s.db, query, args, &rows)
	if err != nil {
		return false, fmt.Errorf("db count has permissions: %w", err)
	}

	if len(rows) == 0 {
		return false, nil
	}

	return rows[0].Count >= 1, nil
}
