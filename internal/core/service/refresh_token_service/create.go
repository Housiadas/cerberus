package refresh_token_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/refresh_token"
	"github.com/google/uuid"
)

// Create adds a new refresh token to the system.
func (c *Service) Create(
	ctx context.Context,
	userID uuid.UUID,
	refreshTokenTTL time.Duration,
) (refresh_token.RefreshToken, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return refresh_token.RefreshToken{}, fmt.Errorf("uuid: %w", err)
	}

	tokenID, err := c.uuidGen.Generate()
	if err != nil {
		return refresh_token.RefreshToken{}, fmt.Errorf("uuid: %w", err)
	}

	now := c.clock.Now()
	tkn := refresh_token.RefreshToken{
		ID:        id,
		UserID:    userID,
		Token:     tokenID.String(),
		CreatedAt: now,
		ExpiresAt: now.UTC().Add(refreshTokenTTL),
		Revoked:   false,
	}

	err = c.storer.Create(ctx, tkn)
	if err != nil {
		return refresh_token.RefreshToken{}, fmt.Errorf("create: %w", err)
	}

	return tkn, nil
}
