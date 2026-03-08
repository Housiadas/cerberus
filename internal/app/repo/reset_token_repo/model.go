package reset_token_repo

import (
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/reset_token"
	"github.com/google/uuid"
)

type tokenDB struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	Used      bool      `db:"used"`
}

func toTokenDB(rt reset_token.ResetToken) tokenDB {
	return tokenDB{
		ID:        rt.ID(),
		UserID:    rt.UserID(),
		Token:     rt.Token(),
		ExpiresAt: rt.ExpiresAt(),
		CreatedAt: rt.CreatedAt(),
		Used:      rt.Used(),
	}
}

func toTokenDomain(db tokenDB) reset_token.ResetToken {
	return reset_token.New(
		db.ID,
		db.UserID,
		db.Token,
		db.ExpiresAt,
		db.CreatedAt,
		db.Used,
	)
}
