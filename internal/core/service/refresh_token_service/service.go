// Package refresh_token_service provides internal access to the domain.
package refresh_token_service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/domain/refresh_token"
	"github.com/Housiadas/cerberus/pkg/logger"
)

// Service manages the set of APIs for user access.
type Service struct {
	log    *logger.Logger
	storer refresh_token.Storer
}

// New constructs a user.User internal API for use.
func New(log *logger.Logger, storer refresh_token.Storer) *Service {
	return &Service{
		log:    log,
		storer: storer,
	}
}

// Create adds a new refresh token to the system.
func (c *Service) Create(ctx context.Context, userID uuid.UUID, ttl time.Duration) (refresh_token.RefreshToken, error) {
	now := time.Now()
	tokenID := uuid.New()
	tkn := refresh_token.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     tokenID.String(),
		CreatedAt: now,
		ExpiresAt: now.UTC().Add(ttl),
		Revoked:   false,
	}

	if err := c.storer.Create(ctx, tkn); err != nil {
		return refresh_token.RefreshToken{}, fmt.Errorf("create: %w", err)
	}

	return tkn, nil
}

// Delete removes the specified user.
func (c *Service) Delete(ctx context.Context, tkn refresh_token.RefreshToken) error {
	if err := c.storer.Delete(ctx, tkn); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// QueryByID finds the user by the specified ID.
func (c *Service) QueryByID(ctx context.Context, tokenID uuid.UUID) (refresh_token.RefreshToken, error) {
	tkn, err := c.storer.QueryByID(ctx, tokenID)
	if err != nil {
		return refresh_token.RefreshToken{}, fmt.Errorf("query: userID[%s]: %w", tokenID, err)
	}

	return tkn, nil
}
