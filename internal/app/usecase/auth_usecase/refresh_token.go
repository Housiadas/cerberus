package auth_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/errs"
)

func (u *UseCase) createRefreshToken(_ context.Context, userID string) (string, error) {
	return "", nil
}

func (u *UseCase) RefreshAccessToken(ctx context.Context, authRefresh RefreshTokenReq) (Token, error) {
	// Validate refresh token
	claims, err := u.Validate(ctx, authRefresh.Token)
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
	aToken, err := u.generateAccessToken(ctx, claims.Subject)
	if err != nil {
		return Token{}, fmt.Errorf("generate access token: %w", err)
	}

	rToken, err := u.createRefreshToken(ctx, claims.Subject)
	if err != nil {
		return Token{}, fmt.Errorf("create refresh token: %w", err)
	}

	return Token{
		AccessToken:  aToken.token,
		RefreshToken: rToken,
		ExpiresIn:    aToken.expiresIn,
	}, nil
}
