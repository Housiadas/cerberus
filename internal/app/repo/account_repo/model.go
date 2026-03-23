package account_repo

import (
	"database/sql"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/account"
	"github.com/google/uuid"
)

type accountDB struct {
	ID               uuid.UUID      `db:"id"`
	Name             string         `db:"name"`
	StripeCustomerID sql.NullString `db:"stripe_customer_id"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
	DeletedAt        sql.NullTime   `db:"deleted_at"`
}

func toAccountDB(acc account.Account) accountDB {
	return accountDB{
		ID:               acc.ID(),
		Name:             acc.Name(),
		StripeCustomerID: acc.StripeCustomerID(),
		CreatedAt:        acc.CreatedAt().UTC(),
		UpdatedAt:        acc.UpdatedAt().UTC(),
		DeletedAt:        toNullTime(acc.DeletedAt()),
	}
}

func toAccountDomain(db accountDB) account.Account {
	return account.New(
		db.ID,
		db.Name,
		db.StripeCustomerID,
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
		fromNullTime(db.DeletedAt),
	)
}

func toAccountsDomain(dbs []accountDB) []account.Account {
	accs := make([]account.Account, len(dbs))
	for i, db := range dbs {
		accs[i] = toAccountDomain(db)
	}

	return accs
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: t.UTC(), Valid: true}
}

func fromNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}

	t := nt.Time.In(time.UTC)

	return &t
}
