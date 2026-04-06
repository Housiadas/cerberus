package auth

import (
	"context"
	"fmt"
	"time"

	errs2 "github.com/Housiadas/cerberus/internal/errs"
)

func (s *Service) RefreshAccessToken(
	ctx context.Context,
	authRefresh RefreshTokenReq,
) (Token, error) {
	// Retrieve the refresh token
	rToken, err := s.refreshTokenService.QueryByToken(ctx, authRefresh.Token)
	if err != nil {
		return Token{}, fmt.Errorf("query by token: %w", err)
	}

	if rToken.Revoked() {
		return Token{}, errs2.New(errs2.InvalidArgument, errs2.CodeInvalidToken, ErrInvalidToken)
	}

	// Check if the token has expired
	if time.Now().UTC().After(rToken.ExpiresAt()) {
		return Token{}, errs2.New(errs2.InvalidArgument, errs2.CodeExpiredToken, ErrExpiredToken)
	}

	// Get the user
	usr, err := s.userService.QueryByID(ctx, rToken.UserID())
	if err != nil {
		return Token{}, fmt.Errorf("query by id: %w", err)
	}

	// Generate a new access token
	aToken, err := s.GenerateAccessToken(ctx, usr.ID().String())
	if err != nil {
		return Token{}, err
	}

	return Token{
		AccessToken:  aToken.Token,
		RefreshToken: rToken.Token(),
		ExpiresIn:    aToken.ExpiresIn,
	}, nil
}
