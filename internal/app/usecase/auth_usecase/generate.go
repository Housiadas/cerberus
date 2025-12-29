package auth_usecase

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/pkg/errs"
)

type accessToken struct {
	token     string
	expiresIn int64
}

func (u *UseCase) generateAccessToken(ctx context.Context, userID string) (accessToken, error) {
	// get user roles name
	roles, err := u.userRolesUsecase.GetUserRolesNames(ctx, userID)
	if err != nil {
		return accessToken{}, errs.Errorf(errs.NotFound, "roles not found: %s", err)
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
	accessTokenID := uuid.New().String()
	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    u.Issuer(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.UTC().Add(accessTokenTTL)),
			Audience:  []string{u.Issuer()},
		},
		TokenID: accessTokenID,
		Roles:   roles,
	}

	aToken := jwt.NewWithClaims(u.method, accessClaims)
	accessTokenString, err := aToken.SignedString(accessTokenSecret)
	if err != nil {
		return accessToken{}, errs.Errorf(errs.InvalidArgument, "failed to sign access token: %s", err)
	}

	expirationDate, err := aToken.Claims.GetExpirationTime()
	if err != nil {
		return accessToken{}, errs.Errorf(errs.InvalidArgument, "expiration time: %s", err)
	}

	return accessToken{
		token:     accessTokenString,
		expiresIn: expirationDate.Unix(),
	}, nil
}
