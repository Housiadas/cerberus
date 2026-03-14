package payment_method_repo

import (
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/payment_method"
	"github.com/google/uuid"
)

type paymentMethodDB struct {
	ID        uuid.UUID `db:"id"`
	AccountID uuid.UUID `db:"account_id"`
	Type      string    `db:"type"`
	Details   string    `db:"details"`
	IsDefault bool      `db:"is_default"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func toPaymentMethodDB(pm payment_method.PaymentMethod) paymentMethodDB {
	return paymentMethodDB{
		ID:        pm.ID(),
		AccountID: pm.AccountID(),
		Type:      pm.Type().String(),
		Details:   pm.Details(),
		IsDefault: pm.IsDefault(),
		CreatedAt: pm.CreatedAt().UTC(),
		UpdatedAt: pm.UpdatedAt().UTC(),
	}
}

func toDomain(db paymentMethodDB) (payment_method.PaymentMethod, error) {
	typ, err := payment_method.Parse(db.Type)
	if err != nil {
		return payment_method.PaymentMethod{}, fmt.Errorf("parse type: %w", err)
	}

	return payment_method.New(
		db.ID,
		db.AccountID,
		typ,
		db.Details,
		db.IsDefault,
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	), nil
}

func toSliceDomain(dbs []paymentMethodDB) ([]payment_method.PaymentMethod, error) {
	pms := make([]payment_method.PaymentMethod, len(dbs))

	for i, db := range dbs {
		var err error

		pms[i], err = toDomain(db)
		if err != nil {
			return nil, fmt.Errorf("to payment methods domain error: %w", err)
		}
	}

	return pms, nil
}
