package refresh_token_usecase

import (
	"context"
	"time"

	"github.com/Housiadas/cerberus/internal/core/service/refresh_token_service"
	"github.com/Housiadas/cerberus/pkg/web/errs"
	"github.com/google/uuid"
)

// UseCase manages the set of cli layer api functions for the user core.
type UseCase struct {
	refreshTokenService *refresh_token_service.Service
}

// NewUseCase constructs a user cli API for use.
func NewUseCase(refreshTokenService *refresh_token_service.Service) *UseCase {
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
		return RefreshToken{}, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	tkn, err := uc.refreshTokenService.Create(ctx, userUUID, refreshTokenTTL)
	if err != nil {
		return RefreshToken{}, errs.Errorf(
			errs.Internal,
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
		return RefreshToken{}, errs.Errorf(errs.Internal, "query by token: [%+v]: %s", tkn, err)
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
		return errs.Errorf(errs.Internal, "revoke issue: [%+v]: %s", tkn, err)
	}

	return nil
}
