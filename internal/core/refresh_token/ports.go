package refresh_token

import (
	"context"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/google/uuid"
)

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}

// storer interface declares the behavior this package needs to persist and retrieve data.
type storer interface {
	CreateRefreshToken(ctx context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error)
	GetRefreshTokenByToken(ctx context.Context, token string) (db.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}
