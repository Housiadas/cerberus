package context

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/domain/permission"
)

const (
	permissionIDKey ctxKey = "permissionIDKey"
	permissionKey   ctxKey = "permissionKey"
)

func SetPermissionID(ctx context.Context, permissionID uuid.UUID) context.Context {
	return context.WithValue(ctx, permissionIDKey, permissionID)
}

// GetPermissionID returns the permission id from the context.
func GetPermissionID(ctx context.Context) (uuid.UUID, error) {
	v, ok := ctx.Value(permissionIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("permission id not found in context")
	}

	return v, nil
}

func SetPermission(ctx context.Context, p permission.Permission) context.Context {
	return context.WithValue(ctx, permissionKey, p)
}

// GetPermission returns the permission from the context.
func GetPermission(ctx context.Context) (permission.Permission, error) {
	v, ok := ctx.Value(permissionKey).(permission.Permission)
	if !ok {
		return permission.Permission{}, errors.New("permission not found in context")
	}

	return v, nil
}
