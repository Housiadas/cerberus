package apitest

import (
	"github.com/Housiadas/cerberus/internal/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/refresh_token_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/logger"
)

type Usecase struct {
	Auth *auth_usecase.UseCase
}

func newUseCase(
	log *logger.Service,
	core *Core,
	accessTokenSecret []byte,
	serviceName string,
) *Usecase {
	userUsecase := user_usecase.NewUseCase(core.User, core.Outbox)
	refreshTokenUsecase := refresh_token_usecase.NewUseCase(core.RefreshToken)
	authUsecase := auth_usecase.NewUseCase(auth_usecase.Config{
		Issuer:              serviceName,
		AccessTokenSecret:   accessTokenSecret,
		Log:                 log,
		UserUsecase:         userUsecase,
		RefreshTokenUsecase: refreshTokenUsecase,
	})

	return &Usecase{
		Auth: authUsecase,
	}
}
