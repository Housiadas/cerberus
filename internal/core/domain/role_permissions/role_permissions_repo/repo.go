// Package role_permissions_repo contains write operations for the role_permissions table.
package role_permissions_repo

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	rp "github.com/Housiadas/cerberus/internal/core/domain/role_permissions"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed query/role_permission_add.sql
	rolePermissionAddSQL string
	//go:embed query/role_permission_remove.sql
	rolePermissionRemoveSQL string
)

// Store manages write operations on the role_permissions table.
type Store struct {
	log logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log logger.Logger, db *sqlx.DB) *Store {
	return &Store{log: log, db: db}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (rp.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("role_permissions transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Add inserts a role-permission relationship.
func (s *Store) Add(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error {
	data := struct {
		RoleID       uuid.UUID `db:"role_id"`
		PermissionID uuid.UUID `db:"permission_id"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
	}{
		RoleID:       roleID,
		PermissionID: permissionID,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	err := pgsql.NamedExecContext(ctx, s.log, s.db, rolePermissionAddSQL, data)
	if err != nil {
		return fmt.Errorf("role_permissions add: %w", err)
	}

	return nil
}

// Remove a role-permission relationship.
func (s *Store) Remove(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error {
	data := struct {
		RoleID       uuid.UUID `db:"role_id"`
		PermissionID uuid.UUID `db:"permission_id"`
	}{
		RoleID:       roleID,
		PermissionID: permissionID,
	}

	err := pgsql.NamedExecContext(ctx, s.log, s.db, rolePermissionRemoveSQL, data)
	if err != nil {
		return fmt.Errorf("role_permissions remove: %w", err)
	}

	return nil
}
