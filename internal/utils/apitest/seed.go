package apitest

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/seed_role.sql
	seedRoleSQL string
	//go:embed query/seed_user_role.sql
	seedUserRoleSQL string
	//go:embed query/seed_permission.sql
	seedPermissionSQL string
	//go:embed query/seed_role_permission.sql
	seedRolePermissionSQL string
)

// SeedRole inserts a role into the database and returns its generated UUID.
func SeedRole(ctx context.Context, db *sqlx.DB, name string) (uuid.UUID, error) {
	id := uuid.New()
	now := time.Now()

	_, err := db.ExecContext(ctx, seedRoleSQL, id, name, now, now)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("seed role: name[%s]: %w", name, err)
	}

	return id, nil
}

// SeedPermission inserts a permission into the database and returns its generated UUID.
func SeedPermission(ctx context.Context, db *sqlx.DB, name string) (uuid.UUID, error) {
	id := uuid.New()
	now := time.Now()

	_, err := db.ExecContext(ctx, seedPermissionSQL, id, name, now, now)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("seed permission: name[%s]: %w", name, err)
	}

	return id, nil
}

// SeedRolePermission inserts a role_permission association into the database.
func SeedRolePermission(ctx context.Context, db *sqlx.DB, roleID, permissionID uuid.UUID) error {
	now := time.Now()

	_, err := db.ExecContext(ctx, seedRolePermissionSQL, roleID, permissionID, now, now)
	if err != nil {
		return fmt.Errorf(
			"seed role_permission: roleID[%s] permissionID[%s]: %w",
			roleID,
			permissionID,
			err,
		)
	}

	return nil
}

// SeedUserRole inserts a user_role association into the database.
func SeedUserRole(ctx context.Context, db *sqlx.DB, userID, roleID uuid.UUID) error {
	now := time.Now()

	_, err := db.ExecContext(ctx, seedUserRoleSQL, userID, roleID, now, now)
	if err != nil {
		return fmt.Errorf("seed user_role: userID[%s] roleID[%s]: %w", userID, roleID, err)
	}

	return nil
}
