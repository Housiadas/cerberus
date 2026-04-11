package apitest

import (
	"context"
	"fmt"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// SeedData represents users for api tests.
type SeedData struct {
	Users  []User
	Admins []User
}

// SeedRole inserts a role into the database and returns its generated UUID.
func SeedRole(
	ctx context.Context,
	store *db.Store,
	name string,
) (uuid.UUID, error) {
	id := uuid.New()
	now := time.Now()

	rol, err := store.CreateRole(ctx, db.CreateRoleParams{
		ID:        id,
		Name:      name,
		CreatedAt: pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: now, Valid: true},
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("seed role: name[%s]: %w", name, err)
	}

	return rol.ID, nil
}

// SeedPermission inserts a permission into the database and returns its generated UUID.
func SeedPermission(
	ctx context.Context,
	store *db.Store,
	name string,
) (uuid.UUID, error) {
	id := uuid.New()
	now := time.Now()

	perm, err := store.CreatePermission(ctx, db.CreatePermissionParams{
		ID:        id,
		Name:      name,
		CreatedAt: pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: now, Valid: true},
	})

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("seed permission: name[%s]: %w", name, err)
	}

	return perm.ID, nil
}

// SeedRolePermission inserts a role_permission association into the database.
func SeedRolePermission(
	ctx context.Context,
	store *db.Store,
	roleID, permissionID uuid.UUID,
) error {
	now := time.Now()

	_, err := store.CreateRolePermission(ctx, db.CreateRolePermissionParams{
		RoleID:       roleID,
		PermissionID: permissionID,
		CreatedAt:    pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:    pgtype.Timestamp{Time: now, Valid: true},
	})
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
func SeedUserRole(
	ctx context.Context,
	store *db.Store,
	userID, roleID uuid.UUID,
) error {
	now := time.Now()

	_, err := store.CreateUserRole(ctx, db.CreateUserRoleParams{
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: now, Valid: true},
	})
	if err != nil {
		return fmt.Errorf(
			"seed user_role: userID[%s] roleID[%s]: %w",
			userID,
			roleID,
			err,
		)
	}

	return nil
}
