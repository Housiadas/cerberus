package auth_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/errs"
)

func (u *UseCase) Login(ctx context.Context, authLogin AuthLogin) (Token, error) {
	authUsr := user_usecase.AuthenticateUser{
		Email:    authLogin.Email,
		Password: authLogin.Password,
	}

	usr, err := u.userUsecase.Authenticate(ctx, authUsr)
	if err != nil {
		return Token{}, errs.New(errs.Unauthenticated, err)
	}

	tokenPair, err := u.Generate(ctx, usr.ID)
	if err != nil {
		return Token{}, fmt.Errorf("generate token pair: %w", err)
	}

	return tokenPair, nil
}
