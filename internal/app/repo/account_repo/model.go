package account_repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/account"
	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/google/uuid"
)

type accountDB struct {
	ID        uuid.UUID    `db:"id"`
	Name      string       `db:"name"`
	Type      string       `db:"type"`
	Enabled   bool         `db:"enabled"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func toAccountDB(acct account.Account) accountDB {
	return accountDB{
		ID:        acct.ID(),
		Name:      acct.Name().String(),
		Type:      acct.Type().String(),
		Enabled:   acct.Enabled(),
		CreatedAt: acct.CreatedAt().UTC(),
		UpdatedAt: acct.UpdatedAt().UTC(),
		DeletedAt: toNullTime(acct.DeletedAt()),
	}
}

func toAccountDomain(db accountDB) (account.Account, error) {
	nme, err := name.Parse(db.Name)
	if err != nil {
		return account.Account{}, fmt.Errorf("parse name: %w", err)
	}

	typ, err := account.Parse(db.Type)
	if err != nil {
		return account.Account{}, fmt.Errorf("parse type: %w", err)
	}

	return account.New(
		db.ID,
		nme,
		typ,
		db.Enabled,
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
		fromNullTime(db.DeletedAt),
	), nil
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

func toAccountsDomain(dbs []accountDB) ([]account.Account, error) {
	items := make([]account.Account, len(dbs))

	for i, db := range dbs {
		var err error

		items[i], err = toAccountDomain(db)
		if err != nil {
			return nil, fmt.Errorf("to accounts domain error: %w", err)
		}
	}

	return items, nil
}
