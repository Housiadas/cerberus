package account

import (
	"database/sql"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toCreateAccountParams(id uuid.UUID, na NewAccount, now time.Time) db.CreateAccountParams {
	return db.CreateAccountParams{
		ID:        id,
		Name:      na.Name,
		CreatedAt: pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: now, Valid: true},
	}
}

func toUpdateAccountParams(acc Account) db.UpdateAccountParams {
	return db.UpdateAccountParams{
		ID:   acc.ID(),
		Name: pgtype.Text{String: acc.Name(), Valid: true},
		StripeCustomerID: pgtype.Text{
			String: acc.StripeCustomerID().String,
			Valid:  acc.StripeCustomerID().Valid,
		},
		UpdatedAt: pgtype.Timestamp{Time: acc.UpdatedAt(), Valid: true},
	}
}

func toDBQueryFilter(f QueryFilter) db.AccountQueryFilter {
	return db.AccountQueryFilter{
		ID:               f.ID,
		Name:             f.Name,
		StripeCustomerID: f.StripeCustomerID,
	}
}

func toDomainAccount(a db.Account) Account {
	var deletedAt *time.Time
	if a.DeletedAt.Valid {
		deletedAt = &a.DeletedAt.Time
	}

	return New(
		a.ID,
		a.Name,
		sql.NullString{String: a.StripeCustomerID.String, Valid: a.StripeCustomerID.Valid},
		a.CreatedAt.Time,
		a.UpdatedAt.Time,
		deletedAt,
	)
}

func toDomainAccounts(dbAccounts []db.Account) []Account {
	accounts := make([]Account, len(dbAccounts))
	for i, a := range dbAccounts {
		accounts[i] = toDomainAccount(a)
	}

	return accounts
}

func toDomainAccountFromGetByID(a db.GetAccountByIDRow) Account {
	return New(
		a.ID,
		a.Name,
		sql.NullString{String: a.StripeCustomerID.String, Valid: a.StripeCustomerID.Valid},
		a.CreatedAt.Time,
		a.UpdatedAt.Time,
		nil,
	)
}
