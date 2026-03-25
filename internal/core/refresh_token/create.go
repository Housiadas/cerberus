package refresh_token

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Create adds a new refresh token to the system.
func (c *Service) Create(
	ctx context.Context,
	userID uuid.UUID,
	refreshTokenTTL time.Duration,
) (RefreshToken, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return RefreshToken{}, fmt.Errorf("uuid: %w", err)
	}

	tokenID, err := c.uuidGen.Generate()
	if err != nil {
		return RefreshToken{}, fmt.Errorf("uuid: %w", err)
	}

	now := c.clock.Now()
	tkn := New(
		id,
		userID,
		tokenID.String(),
		now.UTC().Add(refreshTokenTTL),
		now,
		false,
	)

	err = c.storer.Create(ctx, tkn)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("create: %w", err)
	}

	return tkn, nil
}
