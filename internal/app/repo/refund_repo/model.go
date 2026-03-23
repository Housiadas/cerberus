package refund_repo

import (
	"database/sql"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/refund"
	"github.com/google/uuid"
)

type refundDB struct {
	ID             uuid.UUID      `db:"id"`
	AccountID      uuid.UUID      `db:"account_id"`
	PaymentID      uuid.UUID      `db:"payment_id"`
	StripeRefundID string         `db:"stripe_refund_id"`
	Amount         int64          `db:"amount"`
	Currency       string         `db:"currency"`
	Status         string         `db:"status"`
	Reason         sql.NullString `db:"reason"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

func toRefundDB(ref refund.Refund) refundDB {
	return refundDB{
		ID:             ref.ID(),
		AccountID:      ref.AccountID(),
		PaymentID:      ref.PaymentID(),
		StripeRefundID: ref.StripeRefundID(),
		Amount:         ref.Amount(),
		Currency:       ref.Currency(),
		Status:         ref.Status(),
		Reason:         toNullString(ref.Reason()),
		CreatedAt:      ref.CreatedAt().UTC(),
		UpdatedAt:      ref.UpdatedAt().UTC(),
	}
}

func toRefundDomain(db refundDB) refund.Refund {
	return refund.New(
		db.ID,
		db.AccountID,
		db.PaymentID,
		db.StripeRefundID,
		db.Amount,
		db.Currency,
		db.Status,
		db.Reason.String,
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	)
}

func toRefundsDomain(dbs []refundDB) []refund.Refund {
	refs := make([]refund.Refund, len(dbs))
	for i, db := range dbs {
		refs[i] = toRefundDomain(db)
	}

	return refs
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}

	return sql.NullString{String: s, Valid: true}
}
