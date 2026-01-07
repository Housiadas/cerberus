package refresh_token_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/refresh_token"
)

func (c *Service) QueryByToken(ctx context.Context, token string) (refresh_token.RefreshToken, error) {
	tkn, err := c.storer.QueryByToken(ctx, token)
	if err != nil {
		return refresh_token.RefreshToken{}, fmt.Errorf("query by token: %w", err)
	}

	return tkn, nil
}
