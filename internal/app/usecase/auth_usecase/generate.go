package auth_usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Generate generates a signed JWT token string representing the user Claims.
func (u *UseCase) Generate(ctx context.Context, userID string) (Token, error) {
	// get user roles name
	roles, err := u.userRolesUsecase.GetUserRolesNames(ctx, userID)
	if err != nil {
		return Token{}, fmt.Errorf("roles not found: %w", err)
	}

	// Generating a token requires defining a set of claims
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	now := time.Now()

	// Generate Access Token
	accessTokenID := uuid.New().String()
	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    u.Issuer(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.UTC().Add(accessTokenExpiry)),
			Audience:  []string{u.Issuer()},
		},
		TokenID: accessTokenID,
		Roles:   roles,
	}

	accessToken := jwt.NewWithClaims(u.method, accessClaims)
	accessTokenString, err := accessToken.SignedString(accessTokenSecret)
	if err != nil {
		return Token{}, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate Refresh Token
	refreshTokenID := uuid.New().String()
	refreshClaims := Claims{
		TokenID: refreshTokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    u.Issuer(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.UTC().Add(refreshTokenExpiry)),
			Audience:  []string{u.Issuer()},
		},
		Roles: roles,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(refreshTokenSecret)
	if err != nil {
		return Token{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// todo: Store refresh token in database

	expirationDate, err := accessToken.Claims.GetExpirationTime()
	if err != nil {
		return Token{}, fmt.Errorf("get expiration time: %w", err)
	}

	return Token{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    expirationDate.Unix(),
	}, nil
}
