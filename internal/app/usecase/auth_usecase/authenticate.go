package auth_usecase

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// Authenticate processes the token to validate the sender's token is valid.
func (u *UseCase) Authenticate(ctx context.Context, jwtUnverified string, tt TokenType) (Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(jwtUnverified, &claims, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Only accept HS256
		if token.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, fmt.Errorf("invalid signing algorithm")
		}
		// choose the token type
		switch tt {
		case AccessToken:
			return u.Secrets.AccessTokenSecret, nil
		case RefreshToken:
			return u.Secrets.RefreshTokenSecret, nil
		default:
			panic(fmt.Errorf("unknown token type: %s", tt))
		}
	})
	if err != nil {
		return Claims{}, fmt.Errorf("error parsing token: %w", err)
	}
	if !token.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}

	if err := u.CheckExpiredToken(claims); err != nil {
		return Claims{}, fmt.Errorf("token expired: %w", err)
	}

	// Check if the token is blacklisted
	//if tokenBlacklist[claims.TokenID] {
	//	return nil, fmt.Errorf("token has been revoked")
	//}

	// Check the database for this user to verify they are still enabled.
	if err := u.isUserEnabled(ctx, claims); err != nil {
		return Claims{}, fmt.Errorf("user not enabled: %w", err)
	}

	return claims, nil
}
