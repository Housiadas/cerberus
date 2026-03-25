package refresh_token_usecase

import (
	"context"
	"time"

	"github.com/Housiadas/cerberus/internal/core/refresh_token"
	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/google/uuid"
)

type UseCase struct {
	refreshTokenService *refresh_token.Service
}

func NewUseCase(refreshTokenService *refresh_token.Service) *UseCase {
	return &UseCase{
		refreshTokenService: refreshTokenService,
	}
}

func (uc *UseCase) Create(
	ctx context.Context,
	userID string,
	refreshTokenTTL time.Duration,
) (RefreshToken, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return RefreshToken{}, errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	tkn, err := uc.refreshTokenService.Create(ctx, userUUID, refreshTokenTTL)
	if err != nil {
		return RefreshToken{}, errs2.Errorf(
			errs2.Internal,
			errs2.CodeInternal,
			"create: refresh_token[%+v]: %s",
			tkn,
			err,
		)
	}

	return toAppToken(tkn), nil
}

func (uc *UseCase) QueryByToken(ctx context.Context, token string) (RefreshToken, error) {
	tkn, err := uc.refreshTokenService.QueryByToken(ctx, token)
	if err != nil {
		return RefreshToken{}, errs2.Errorf(
			errs2.Internal,
			errs2.CodeInternal,
			"query by token: [%+v]: %s",
			tkn,
			err,
		)
	}

	return toAppToken(tkn), nil
}

func (uc *UseCase) Revoke(ctx context.Context, tkn RefreshToken) error {
	coreTkn, err := toCoreToken(tkn)
	if err != nil {
		return err
	}

	err = uc.refreshTokenService.Revoke(ctx, coreTkn)
	if err != nil {
		return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "revoke issue: [%+v]: %s", tkn, err)
	}

	return nil
}
