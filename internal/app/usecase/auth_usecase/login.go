package auth_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/errs"
)

func (u *UseCase) Login(ctx context.Context, authLogin LoginReq) (Token, error) {
	authUsr := user_usecase.AuthenticateUser{
		Email:    authLogin.Email,
		Password: authLogin.Password,
	}

	// Verify email and password
	usr, err := u.userUsecase.Authenticate(ctx, authUsr)
	if err != nil {
		return Token{}, errs.New(errs.Unauthenticated, err)
	}

	// Generate JWT access token
	aToken, err := u.generateAccessToken(ctx, usr.ID)
	if err != nil {
		return Token{}, fmt.Errorf("generate access token: %w", err)
	}

	// Create a refresh token
	tkn, err := u.refreshTokenUsecase.Create(ctx, usr.ID, refreshTokenTTL)
	if err != nil {
		return Token{}, fmt.Errorf("create refresh token: %w", err)
	}

	return Token{
		AccessToken:  aToken.token,
		RefreshToken: tkn.Token,
		ExpiresIn:    aToken.expiresIn,
	}, nil
}
