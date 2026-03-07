package refresh_token_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/refresh_token"
)

// Revoke revokes the specified refresh_token.
func (c *Service) Revoke(ctx context.Context, tkn refresh_token.RefreshToken) error {
	revToken := tkn.WithRevoked(true)

	err := c.storer.Revoke(ctx, revToken)
	if err != nil {
		return fmt.Errorf("revoke: %w", err)
	}

	return nil
}
