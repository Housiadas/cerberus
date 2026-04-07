package auth

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/google/uuid"
)

func (s *Service) Logout(ctx context.Context, userID string, req LogoutReq) error {
	// Retrieve refresh token
	rToken, err := s.refreshTokenService.QueryByToken(ctx, req.Token)
	if err != nil {
		return fmt.Errorf("query by token: %w", err)
	}

	// Check if userID matches
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("parse user uuid: %w", err)
	}

	if rToken.UserID() != userUUID {
		return errs.New(
			errs.Unauthenticated,
			errs.CodeUnauthenticated,
			errs.Errorf(errs.Unauthenticated, errs.CodeUnauthenticated, "invalid user id"),
		)
	}

	// Revoke refresh token
	err = s.refreshTokenService.Revoke(ctx, rToken)
	if err != nil {
		return fmt.Errorf("revoke issue: %w", err)
	}

	return nil
}
