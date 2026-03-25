// Package user_roles_repo contains write operations for the user_roles table.
package user_roles_repo

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	ur "github.com/Housiadas/cerberus/internal/core/user_roles"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed query/user_role_add.sql
	userRoleAddSQL string
	//go:embed query/user_role_remove.sql
	userRoleRemoveSQL string
)

// Store manages write operations on the user_roles table.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (ur.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("user_roles transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Add inserts a user-role relationship.
func (s *Store) Add(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	data := struct {
		UserID    uuid.UUID `db:"user_id"`
		RoleID    uuid.UUID `db:"role_id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}{
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := pgsql.NamedExecContext(ctx, s.log, s.db, userRoleAddSQL, data)
	if err != nil {
		return fmt.Errorf("user_roles add: %w", err)
	}

	return nil
}

// Remove a user-role relationship.
func (s *Store) Remove(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	data := struct {
		UserID uuid.UUID `db:"user_id"`
		RoleID uuid.UUID `db:"role_id"`
	}{
		UserID: userID,
		RoleID: roleID,
	}

	err := pgsql.NamedExecContext(ctx, s.log, s.db, userRoleRemoveSQL, data)
	if err != nil {
		return fmt.Errorf("user_roles remove: %w", err)
	}

	return nil
}
