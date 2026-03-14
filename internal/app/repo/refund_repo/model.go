package refund_repo

import (
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/payment"
	"github.com/Housiadas/cerberus/internal/core/domain/refund"
	"github.com/google/uuid"
)

type refundDB struct {
	ID          uuid.UUID `db:"id"`
	PaymentID   uuid.UUID `db:"payment_id"`
	AmountCents int       `db:"amount_cents"`
	Reason      string    `db:"reason"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func toRefundDB(ref refund.Refund) refundDB {
	return refundDB{
		ID:          ref.ID(),
		PaymentID:   ref.PaymentID(),
		AmountCents: ref.AmountCents(),
		Reason:      ref.Reason(),
		Status:      ref.Status().String(),
		CreatedAt:   ref.CreatedAt().UTC(),
		UpdatedAt:   ref.UpdatedAt().UTC(),
	}
}

func toRefundDomain(db refundDB) (refund.Refund, error) {
	status, err := payment.Parse(db.Status)
	if err != nil {
		return refund.Refund{}, fmt.Errorf("parse status: %w", err)
	}

	return refund.New(
		db.ID,
		db.PaymentID,
		db.AmountCents,
		db.Reason,
		status,
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	), nil
}

func toRefundsDomain(dbs []refundDB) ([]refund.Refund, error) {
	refs := make([]refund.Refund, len(dbs))

	for i, db := range dbs {
		var err error

		refs[i], err = toRefundDomain(db)
		if err != nil {
			return nil, fmt.Errorf("to refunds domain error: %w", err)
		}
	}

	return refs, nil
}
