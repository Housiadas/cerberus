package user_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
)

// Delete removes the specified user.
func (c *Service) Delete(ctx context.Context, usr user.User) error {
	if err := c.storer.Delete(ctx, usr); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
