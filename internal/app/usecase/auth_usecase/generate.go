package auth_usecase

import (
	"context"
	"time"

	"github.com/Housiadas/cerberus/pkg/web/errs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessToken struct {
	Token     string
	ExpiresIn int64
}

func (u *UseCase) GenerateAccessToken(_ context.Context, userID string) (AccessToken, error) {
	// get user roles name
	// roles, err := u.userRolesUsecase.GetUserRolesNames(ctx, userID)
	// if err != nil {
	//	return AccessToken{}, errs.Errorf(errs.NotFound, "roles not found: %s", err)
	//}
	// Generating a Token requires defining a set of claims
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed
	// (allows a Token to be used only once)
	now := time.Now()

	accessTokenID, err := uuid.NewV7()
	if err != nil {
		return AccessToken{}, errs.Errorf(errs.Internal, "uuid v7: %s", err)
	}

	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    u.Issuer(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.UTC().Add(accessTokenTTL)),
			Audience:  []string{u.Issuer()},
		},
		TokenID: accessTokenID.String(),
		// Roles:   roles,
	}

	aToken := jwt.NewWithClaims(u.method, accessClaims)

	accessTokenString, err := aToken.SignedString(accessTokenSecret)
	if err != nil {
		return AccessToken{}, errs.Errorf(
			errs.InvalidArgument,
			"failed to sign access Token: %s",
			err,
		)
	}

	expirationDate, err := aToken.Claims.GetExpirationTime()
	if err != nil {
		return AccessToken{}, errs.Errorf(errs.InvalidArgument, "expiration time: %s", err)
	}

	return AccessToken{
		Token:     accessTokenString,
		ExpiresIn: expirationDate.Unix(),
	}, nil
}
