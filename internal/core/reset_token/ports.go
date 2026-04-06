package reset_token

import "context"

// Storer declares the behaviour this package needs to persist and retrieve data.
type Storer interface {
	Create(ctx context.Context, token ResetToken) error
	Delete(ctx context.Context, token ResetToken) error
	QueryByToken(ctx context.Context, token string) (ResetToken, error)
}
