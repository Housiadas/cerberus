package refresh_token_repo

import (
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/refresh_token"
	"github.com/google/uuid"
)

type tokenDB struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	Revoked   bool      `db:"revoked"`
}

func toTokenDB(rToken refresh_token.RefreshToken) tokenDB {
	return tokenDB{
		ID:        rToken.ID,
		UserID:    rToken.UserID,
		Token:     rToken.Token,
		ExpiresAt: rToken.ExpiresAt,
		CreatedAt: rToken.CreatedAt,
		Revoked:   rToken.Revoked,
	}
}

func toTokenDomain(db tokenDB) (refresh_token.RefreshToken, error) {
	bus := refresh_token.RefreshToken{
		ID:        db.ID,
		UserID:    db.UserID,
		Token:     db.Token,
		ExpiresAt: db.ExpiresAt,
		CreatedAt: db.CreatedAt,
		Revoked:   db.Revoked,
	}

	return bus, nil
}
