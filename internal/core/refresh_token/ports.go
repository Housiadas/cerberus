package refresh_token

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}

// Storer interface declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	Create(ctx context.Context, token RefreshToken) error
	Delete(ctx context.Context, token RefreshToken) error
	Revoke(ctx context.Context, token RefreshToken) error
	QueryByToken(ctx context.Context, token string) (RefreshToken, error)
}
