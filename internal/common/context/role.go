package context

import (
	"context"
	"errors"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
	"github.com/google/uuid"
)

func SetRoleID(ctx context.Context, roleID uuid.UUID) context.Context {
	return context.WithValue(ctx, roleIDKey, roleID)
}

// GetRoleID returns the role id from the context.
func GetRoleID(ctx context.Context) (uuid.UUID, error) {
	v, ok := ctx.Value(roleIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("role id not found in context")
	}

	return v, nil
}

func SetRole(ctx context.Context, usr role.Role) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

// GetRole returns the user from the context.
func GetRole(ctx context.Context) (role.Role, error) {
	v, ok := ctx.Value(userKey).(role.Role)
	if !ok {
		return role.Role{}, errors.New("role not found in context")
	}

	return v, nil
}
