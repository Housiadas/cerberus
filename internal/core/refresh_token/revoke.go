package refresh_token

import (
	"context"
	"fmt"
)

// Revoke revokes the specified refresh_token.
func (c *Service) Revoke(ctx context.Context, tkn RefreshToken) error {
	revToken := tkn.WithRevoked(true)

	err := c.storer.Revoke(ctx, revToken)
	if err != nil {
		return fmt.Errorf("revoke: %w", err)
	}

	return nil
}
