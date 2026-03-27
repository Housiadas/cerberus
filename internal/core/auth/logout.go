package auth

import (
	"context"
	"fmt"

	errs2 "github.com/Housiadas/cerberus/internal/errs"
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
		return errs2.New(
			errs2.Unauthenticated,
			errs2.CodeUnauthenticated,
			errs2.Errorf(errs2.Unauthenticated, errs2.CodeUnauthenticated, "invalid user id"),
		)
	}

	// Revoke refresh token
	err = s.refreshTokenService.Revoke(ctx, rToken)
	if err != nil {
		return fmt.Errorf("revoke issue: %w", err)
	}

	return nil
}
