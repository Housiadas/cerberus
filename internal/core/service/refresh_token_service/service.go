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
func (c *Service) Create(ctx context.Context, userID uuid.UUID, refreshTokenTTL time.Duration) (refresh_token.RefreshToken, error) {
	now := time.Now()
	tokenID := uuid.New()
	tkn := refresh_token.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     tokenID.String(),
		CreatedAt: now,
		ExpiresAt: now.UTC().Add(refreshTokenTTL),
		Revoked:   false,
	}

	if err := c.storer.Create(ctx, tkn); err != nil {
		return refresh_token.RefreshToken{}, fmt.Errorf("create: %w", err)
	}

	return tkn, nil
}

// Delete removes the specified refresh_token.
func (c *Service) Delete(ctx context.Context, tkn refresh_token.RefreshToken) error {
	if err := c.storer.Delete(ctx, tkn); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Revoke revokes the specified refresh_token.
func (c *Service) Revoke(ctx context.Context, tkn refresh_token.RefreshToken) error {
	revToken := refresh_token.RefreshToken{
		ID:        tkn.ID,
		UserID:    tkn.UserID,
		Token:     tkn.Token,
		ExpiresAt: tkn.ExpiresAt,
		CreatedAt: tkn.CreatedAt,
		Revoked:   true,
	}
	if err := c.storer.Revoke(ctx, revToken); err != nil {
		return fmt.Errorf("revoke: %w", err)
	}

	return nil
}

func (c *Service) QueryByToken(ctx context.Context, token string) (refresh_token.RefreshToken, error) {
	tkn, err := c.storer.QueryByToken(ctx, token)
	if err != nil {
		return refresh_token.RefreshToken{}, fmt.Errorf("query by token: %w", err)
	}

	return tkn, nil
}
