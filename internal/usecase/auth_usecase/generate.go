package auth_usecase

import (
	"context"
	"time"

	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessToken struct {
	Token     string
	ExpiresIn int64
}

func (u *UseCase) GenerateAccessToken(ctx context.Context, userID string) (AccessToken, error) {
	perms, err := u.userRolesPermissions.QueryPermissionsByUserID(ctx, userID)
	if err != nil {
		return AccessToken{}, errs2.Errorf(
			errs2.Internal,
			errs2.CodeInternal,
			"query permissions: %s",
			err,
		)
	}

	claimPerms := make([]Permission, len(perms))
	for i, p := range perms {
		claimPerms[i] = Permission{ID: p.ID, Name: p.Name}
	}

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

	accessTokenID, genErr := uuid.NewV7()
	if genErr != nil {
		return AccessToken{}, errs2.Errorf(
			errs2.Internal,
			errs2.CodeInternal,
			"uuid v7: %s",
			genErr,
		)
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
		TokenID:     accessTokenID.String(),
		Permissions: claimPerms,
	}

	aToken := jwt.NewWithClaims(u.method, accessClaims)

	accessTokenString, err := aToken.SignedString(u.secret)
	if err != nil {
		return AccessToken{}, errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeInvalidToken,
			"failed to sign access Token: %s",
			err,
		)
	}

	expirationDate, err := aToken.Claims.GetExpirationTime()
	if err != nil {
		return AccessToken{}, errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeInvalidToken,
			"expiration time: %s",
			err,
		)
	}

	return AccessToken{
		Token:     accessTokenString,
		ExpiresIn: expirationDate.Unix(),
	}, nil
}
