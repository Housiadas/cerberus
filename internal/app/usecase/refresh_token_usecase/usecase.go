package refresh_token_usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/service/refresh_token_service"
	"github.com/Housiadas/cerberus/pkg/errs"
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

// Create adds a new role to the system.
func (uc *UseCase) Create(ctx context.Context, userID string, refreshTokenTTL time.Duration) (RefreshToken, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return RefreshToken{}, errs.Newf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	rol, err := uc.refreshTokenService.Create(ctx, userUUID, refreshTokenTTL)
	if err != nil {
		return RefreshToken{}, errs.Newf(errs.Internal, "create: rol[%+v]: %s", rol, err)
	}

	return toAppToken(rol), nil
}
