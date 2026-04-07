package auth

import (
	"context"
	"time"

	"github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessToken struct {
	Token     string
	ExpiresIn int64
}

func (s *Service) GenerateAccessToken(ctx context.Context, userID string) (AccessToken, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return AccessToken{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	perms, err := s.userRolesPermissions.QueryPermissionsByUserID(ctx, userUUID)
	if err != nil {
		return AccessToken{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"query permissions: %s",
			err,
		)
	}

	claimPerms := make([]Permission, len(perms))
	for i, p := range perms {
		claimPerms[i] = Permission{ID: p.ID.String(), Name: p.Name}
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
		return AccessToken{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"uuid v7: %s",
			genErr,
		)
	}

	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    s.Issuer(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.UTC().Add(accessTokenTTL)),
			Audience:  []string{s.Issuer()},
		},
		TokenID:     accessTokenID.String(),
		Permissions: claimPerms,
	}

	aToken := jwt.NewWithClaims(s.method, accessClaims)

	accessTokenString, err := aToken.SignedString(s.secret)
	if err != nil {
		return AccessToken{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeInvalidToken,
			"failed to sign access Token: %s",
			err,
		)
	}

	expirationDate, err := aToken.Claims.GetExpirationTime()
	if err != nil {
		return AccessToken{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeInvalidToken,
			"expiration time: %s",
			err,
		)
	}

	return AccessToken{
		Token:     accessTokenString,
		ExpiresIn: expirationDate.Unix(),
	}, nil
}
