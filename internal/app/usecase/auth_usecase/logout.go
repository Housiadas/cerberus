package auth_usecase

import (
	"context"

	"github.com/Housiadas/cerberus/pkg/errs"
)

func (u *UseCase) Logout(ctx context.Context, userID string, req LogoutReq) error {
	// Retrieve refresh token
	rToken, err := u.refreshTokenUsecase.QueryByToken(ctx, req.Token)
	if err != nil {
		return err
	}

	// Check if userID matches
	if rToken.UserID != userID {
		return errs.New(errs.Unauthenticated, errs.Errorf(errs.Unauthenticated, "invalid user id"))
	}

	// Revoke refresh token
	err = u.refreshTokenUsecase.Revoke(ctx, rToken)
	if err != nil {
		return err
	}

	return nil
}
