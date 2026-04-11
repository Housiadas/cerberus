package reset_token

import (
	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func toCreateResetTokenParams(rt ResetToken) db.CreateResetTokenParams {
	return db.CreateResetTokenParams{
		ID:        rt.id,
		UserID:    rt.userID,
		Token:     rt.token,
		Used:      rt.used,
		ExpiresAt: pgtype.Timestamp{Time: rt.expiresAt, Valid: true},
		CreatedAt: pgtype.Timestamp{Time: rt.createdAt, Valid: true},
	}
}

func toDomainResetToken(rt db.ResetToken) ResetToken {
	return New(
		rt.ID,
		rt.UserID,
		rt.Token,
		rt.ExpiresAt.Time,
		rt.CreatedAt.Time,
		rt.Used,
	)
}
