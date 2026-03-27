package auth

import (
	"context"
	"fmt"
	"net/mail"
)

func (s *Service) Login(ctx context.Context, authLogin LoginReq) (Token, error) {
	addr, err := mail.ParseAddress(authLogin.Email)
	if err != nil {
		return Token{}, fmt.Errorf("parse email: %w", err)
	}

	// Verify email and password
	usr, err := s.userService.Authenticate(ctx, *addr, authLogin.Password)
	if err != nil {
		return Token{}, fmt.Errorf("authenticate user: %w", err)
	}

	// Generate JWT access token
	aToken, err := s.GenerateAccessToken(ctx, usr.ID().String())
	if err != nil {
		return Token{}, err
	}

	// Create a refresh token
	rToken, err := s.refreshTokenService.Create(ctx, usr.ID(), refreshTokenTTL)
	if err != nil {
		return Token{}, fmt.Errorf("create refresh token: %w", err)
	}

	return Token{
		AccessToken:  aToken.Token,
		RefreshToken: rToken.Token(),
		ExpiresIn:    aToken.ExpiresIn,
	}, nil
}
