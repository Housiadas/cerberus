package auth_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
)

func (u *UseCase) Login(ctx context.Context, authLogin LoginReq) (Token, error) {
	authUsr := user_usecase.AuthenticateUser{
		Email:    authLogin.Email,
		Password: authLogin.Password,
	}

	// Verify email and password
	usr, err := u.userUsecase.Authenticate(ctx, authUsr)
	if err != nil {
		return Token{}, fmt.Errorf("authenticate user: %w", err)
	}

	// Generate JWT access token
	aToken, err := u.GenerateAccessToken(ctx, usr.ID)
	if err != nil {
		return Token{}, err
	}

	// Create a refresh token
	rToken, err := u.refreshTokenUsecase.Create(ctx, usr.ID, refreshTokenTTL)
	if err != nil {
		return Token{}, fmt.Errorf("create refresh token: %w", err)
	}

	return Token{
		AccessToken:  aToken.Token,
		RefreshToken: rToken.Token,
		ExpiresIn:    aToken.ExpiresIn,
	}, nil
}
