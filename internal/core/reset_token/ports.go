package reset_token

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

// Storer declares the behavior this package needs to persist and retrieve data.
type Storer interface {
	Create(ctx context.Context, token ResetToken) error
	Delete(ctx context.Context, token ResetToken) error
	QueryByToken(ctx context.Context, token string) (ResetToken, error)
}
