package billing_address_repo

import (
	"database/sql"
	"time"

	"github.com/Housiadas/cerberus/internal/core/billing_address"
	"github.com/google/uuid"
)

type billingAddressDB struct {
	ID         uuid.UUID      `db:"id"`
	AccountID  uuid.UUID      `db:"account_id"`
	Line1      string         `db:"line1"`
	Line2      sql.NullString `db:"line2"`
	City       string         `db:"city"`
	State      sql.NullString `db:"state"`
	PostalCode string         `db:"postal_code"`
	Country    string         `db:"country"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
}

func toBillingAddressDB(addr billing_address.BillingAddress) billingAddressDB {
	return billingAddressDB{
		ID:         addr.ID(),
		AccountID:  addr.AccountID(),
		Line1:      addr.Line1(),
		Line2:      toNullString(addr.Line2()),
		City:       addr.City(),
		State:      toNullString(addr.State()),
		PostalCode: addr.PostalCode(),
		Country:    addr.Country(),
		CreatedAt:  addr.CreatedAt().UTC(),
		UpdatedAt:  addr.UpdatedAt().UTC(),
	}
}

func toBillingAddressDomain(db billingAddressDB) billing_address.BillingAddress {
	return billing_address.New(
		db.ID,
		db.AccountID,
		db.Line1,
		db.Line2.String,
		db.City,
		db.State.String,
		db.PostalCode,
		db.Country,
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
	)
}

func toBillingAddressesDomain(dbs []billingAddressDB) []billing_address.BillingAddress {
	addrs := make([]billing_address.BillingAddress, len(dbs))
	for i, db := range dbs {
		addrs[i] = toBillingAddressDomain(db)
	}

	return addrs
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}

	return sql.NullString{String: s, Valid: true}
}
