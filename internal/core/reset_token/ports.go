package reset_token

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

// storer declares the behavior this package needs to persist and retrieve data.
type storer interface {
	CreateResetToken(ctx context.Context, arg db.CreateResetTokenParams) (db.ResetToken, error)
	DeleteResetToken(ctx context.Context, id uuid.UUID) error
	GetResetTokenByToken(ctx context.Context, token string) (db.ResetToken, error)
}
