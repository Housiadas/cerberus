package auth_usecase

import (
	"context"
	"fmt"

	errs2 "github.com/Housiadas/cerberus/internal/errs"
)

func (u *UseCase) Logout(ctx context.Context, userID string, req LogoutReq) error {
	// Retrieve refresh token
	rToken, err := u.refreshTokenUsecase.QueryByToken(ctx, req.Token)
	if err != nil {
		return fmt.Errorf("query by token: %w", err)
	}

	// Check if userID matches
	if rToken.UserID != userID {
		return errs2.New(
			errs2.Unauthenticated,
			errs2.CodeUnauthenticated,
			errs2.Errorf(errs2.Unauthenticated, errs2.CodeUnauthenticated, "invalid user id"),
		)
	}

	// Revoke refresh token
	err = u.refreshTokenUsecase.Revoke(ctx, rToken)
	if err != nil {
		return fmt.Errorf("revoke issue: %w", err)
	}

	return nil
}
