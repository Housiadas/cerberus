package billing_address_repo

import (
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/billing_address"
	"github.com/Housiadas/cerberus/internal/core/domain/country"
	"github.com/google/uuid"
)

type billingAddressDB struct {
	ID         uuid.UUID `db:"id"`
	AccountID  uuid.UUID `db:"account_id"`
	Line1      string    `db:"line1"`
	Line2      string    `db:"line2"`
	City       string    `db:"city"`
	State      string    `db:"state"`
	PostalCode string    `db:"postal_code"`
	Country    string    `db:"country"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func toBillingAddressDB(ba billing_address.BillingAddress) billingAddressDB {
	return billingAddressDB{
		ID:         ba.ID(),
		AccountID:  ba.AccountID(),
		Line1:      ba.Line1(),
		Line2:      ba.Line2(),
		City:       ba.City(),
		State:      ba.State(),
		PostalCode: ba.PostalCode(),
		Country:    ba.Country().String(),
		CreatedAt:  ba.CreatedAt().UTC(),
		UpdatedAt:  ba.UpdatedAt().UTC(),
	}
}

func toBillingAddressDomain(db billingAddressDB) (billing_address.BillingAddress, error) {
	cntry, err := country.Parse(db.Country)
	if err != nil {
		return billing_address.BillingAddress{}, fmt.Errorf("parse country: %w", err)
	}

	return billing_address.New(
		db.ID,
		db.AccountID,
		db.Line1,
		db.Line2,
		db.City,
		db.State,
		db.PostalCode,
		cntry,
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	), nil
}
