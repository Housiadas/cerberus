package auth_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/errs"
)

func (u *UseCase) Refresh(ctx context.Context, authRefresh AuthRefreshToken) (Token, error) {
	// Validate refresh token
	claims, err := u.Authenticate(ctx, authRefresh.Token, RefreshToken)
	if err != nil {
		return Token{}, errs.New(errs.Unauthenticated, err)
	}

	// Check if a refresh token exists in store
	// Check if the refresh token exists in store
	//if !refreshTokenStore[claims.TokenID] {
	//	return Token{}, fmt.Errorf("refresh token not exists: %w", err)
	//}

	// Revoke old refresh token

	// Generate a new token pair
	tokenPair, err := u.Generate(ctx, claims.Subject)
	if err != nil {
		return Token{}, fmt.Errorf("generate token pair: %w", err)
	}

	return tokenPair, nil
}
