package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/golang-jwt/jwt/v5"
)

// Validate processes for the JWT token.
func (s *Service) Validate(ctx context.Context, jwtUnverified string) (Claims, error) {
	var claims Claims

	token, err := jwt.ParseWithClaims(jwtUnverified, &claims, func(token *jwt.Token) (any, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errs.New(errs.InvalidArgument, errs.CodeInvalidToken, ErrInvalidToken)
		}
		// Only accept HS256
		if token.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, errs.New(errs.InvalidArgument, errs.CodeInvalidToken, ErrInvalidToken)
		}

		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Claims{}, fmt.Errorf("token expired: %w", err)
		}

		return Claims{}, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return Claims{}, errs.New(errs.InvalidArgument, errs.CodeInvalidToken, ErrInvalidToken)
	}

	err = s.CheckExpiredToken(claims)
	if err != nil {
		return Claims{}, fmt.Errorf("token expired: %w", err)
	}

	// Check the database for this user to verify they are still enabled.
	err = s.isUserEnabled(ctx, claims)
	if err != nil {
		if errors.Is(err, ErrUserDisabled) {
			return Claims{}, errs.New(
				errs.Unauthenticated,
				errs.CodeUserDisabled,
				ErrUserDisabled,
			)
		}

		return Claims{}, err
	}

	return claims, nil
}
