package user

import (
	"net/mail"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toCreateUserParams(
	id uuid.UUID,
	usr User,
	now time.Time,
) db.CreateUserParams {
	params := db.CreateUserParams{
		ID:           id,
		Name:         usr.Name().String(),
		Email:        usr.Email().Address,
		PasswordHash: string(usr.PasswordHash()),
		Department: pgtype.Text{
			String: usr.Department().String(),
			Valid:  usr.Department().Valid(),
		},
		Enabled:   usr.Enabled(),
		CreatedAt: pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: now, Valid: true},
	}

	if usr.AccountID() != nil {
		params.AccountID = pgtype.UUID{Bytes: *usr.AccountID(), Valid: true}
	}

	return params
}

func toUpdateUserParams(usr User) db.UpdateUserParams {
	return db.UpdateUserParams{
		ID:           usr.ID(),
		Name:         pgtype.Text{String: usr.Name().String(), Valid: true},
		Email:        pgtype.Text{String: usr.Email().Address, Valid: true},
		PasswordHash: pgtype.Text{String: string(usr.PasswordHash()), Valid: true},
		Enabled:      pgtype.Bool{Bool: usr.Enabled(), Valid: true},
		UpdatedAt:    pgtype.Timestamp{Time: usr.UpdatedAt(), Valid: true},
	}
}

func toDBQueryFilter(f QueryFilter) db.UserQueryFilter {
	dbf := db.UserQueryFilter{
		ID:               f.ID,
		StartCreatedDate: f.StartCreatedDate,
		EndCreatedDate:   f.EndCreatedDate,
	}

	if f.Name != nil {
		dbf.Name = new(f.Name.String())
	}

	if f.Email != nil {
		dbf.Email = new(f.Email.Address)
	}

	return dbf
}

func toDomainUser(u db.User) User {
	var accountID *uuid.UUID
	if u.AccountID.Valid {
		accountID = (*uuid.UUID)(&u.AccountID.Bytes)
	}

	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		deletedAt = &u.DeletedAt.Time
	}

	return New(
		u.ID,
		name.MustParse(u.Name),
		mail.Address{Address: u.Email},
		[]byte(u.PasswordHash),
		toNullName(u.Department),
		u.Enabled,
		accountID,
		u.CreatedAt.Time,
		u.UpdatedAt.Time,
		deletedAt,
	)
}

func toDomainUsers(dbUsers []db.User) []User {
	users := make([]User, len(dbUsers))
	for i, u := range dbUsers {
		users[i] = toDomainUser(u)
	}

	return users
}

func toNullName(t pgtype.Text) name.Null {
	if !t.Valid {
		return name.Null{}
	}

	return name.MustParseNull(t.String)
}
