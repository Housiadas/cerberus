package refresh_token

import (
	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func toCreateRefreshTokenParams(rt RefreshToken) db.CreateRefreshTokenParams {
	return db.CreateRefreshTokenParams{
		ID:        rt.id,
		UserID:    rt.userID,
		Token:     rt.token,
		ExpiresAt: pgtype.Timestamp{Time: rt.expiresAt, Valid: true},
		CreatedAt: pgtype.Timestamp{Time: rt.createdAt, Valid: true},
		Revoked:   rt.revoked,
	}
}

func toDomainRefreshToken(rt db.RefreshToken) RefreshToken {
	return New(
		rt.ID,
		rt.UserID,
		rt.Token,
		rt.ExpiresAt.Time,
		rt.CreatedAt.Time,
		rt.Revoked,
	)
}
