package refresh_token

import (
	"context"
	"fmt"
)

func (c *Service) QueryByToken(
	ctx context.Context,
	token string,
) (RefreshToken, error) {
	tkn, err := c.storer.QueryByToken(ctx, token)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("query by token: %w", err)
	}

	return tkn, nil
}
