package context

import (
	"context"

	"github.com/Housiadas/cerberus/internal/app/usecase/auth_usecase"
)

const (
	claimsKey ctxKey = "claimsKey"
)

func SetClaims(ctx context.Context, claims auth_usecase.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// GetClaims returns the claims from the context.
func GetClaims(ctx context.Context) auth_usecase.Claims {
	v, ok := ctx.Value(claimsKey).(auth_usecase.Claims)
	if !ok {
		return auth_usecase.Claims{}
	}
	return v
}
