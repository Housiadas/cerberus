package auth_usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/pkg/errs"
)

func (u *UseCase) RefreshAccessToken(ctx context.Context, authRefresh RefreshTokenReq) (Token, error) {
	// Retrieve the refresh token
	rToken, err := u.refreshTokenUsecase.QueryByToken(ctx, authRefresh.Token)
	if err != nil {
		return Token{}, err
	}

	// Check if the token is valid
	if rToken.Revoked {
		return Token{}, errs.New(errs.InvalidArgument, ErrInvalidToken)
	}

	// Check if the token has expired
	expiresAt, err := time.Parse(time.RFC3339, rToken.ExpiresAt)
	if err != nil {
		return Token{}, fmt.Errorf("parse time: %w", err)
	}
	if time.Now().UTC().After(expiresAt) {
		return Token{}, errs.New(errs.InvalidArgument, ErrExpiredToken)
	}

	// Get the user
	usr, err := u.userUsecase.QueryByID(ctx, rToken.UserID)
	if err != nil {
		return Token{}, err
	}

	// Generate a new token pair
	aToken, err := u.generateAccessToken(ctx, usr.ID)
	if err != nil {
		return Token{}, err
	}

	return Token{
		AccessToken:  aToken.token,
		RefreshToken: rToken.Token,
		ExpiresIn:    aToken.expiresIn,
	}, nil
}
