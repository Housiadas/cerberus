package context

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
)

const (
	userIDKey ctxKey = "userIDKey"
	userKey   ctxKey = "userKey"
)

func SetUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID returns the user id from the context.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("user id not found in context")
	}

	return v, nil
}

func SetUser(ctx context.Context, usr user.User) context.Context {
	return context.WithValue(ctx, userKey, usr)
}

// GetUser returns the user from the context.
func GetUser(ctx context.Context) (user.User, error) {
	v, ok := ctx.Value(userKey).(user.User)
	if !ok {
		return user.User{}, errors.New("user not found in context")
	}

	return v, nil
}
