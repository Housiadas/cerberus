package refresh_token

import (
	"context"

	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	Create(ctx context.Context, token RefreshToken) error
	//Update(ctx context.Context, token RefreshToken) error
	Delete(ctx context.Context, token RefreshToken) error
	QueryByID(ctx context.Context, tokenID uuid.UUID) (RefreshToken, error)
}
