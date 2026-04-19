package user

import (
	"fmt"
	"net/mail"
	"strconv"
	"time"

	db "github.com/Housiadas/cerberus/db/sqlc"
	"github.com/Housiadas/cerberus/internal/core/user/user_search"
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
		Department: pgtype.Text{
			String: usr.Department().String(),
			Valid:  usr.Department().Valid(),
		},
		Enabled:   pgtype.Bool{Bool: usr.Enabled(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: usr.UpdatedAt(), Valid: true},
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

func toNullName(t pgtype.Text) name.Null {
	if !t.Valid {
		return name.Null{}
	}

	return name.MustParseNull(t.String)
}

func toUserFromDoc(doc user_search.UserDoc) (User, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return User{}, fmt.Errorf("parsing user id: %w", err)
	}

	nm, err := name.Parse(doc.Name)
	if err != nil {
		return User{}, fmt.Errorf("parsing user name: %w", err)
	}

	dept := name.Null{}
	if doc.Department != "" {
		dept, err = name.ParseNull(doc.Department)
		if err != nil {
			return User{}, fmt.Errorf("parsing department: %w", err)
		}
	}

	createdAt, err := parseESDate(doc.CreatedAt)
	if err != nil {
		return User{}, fmt.Errorf("parsing createdAt: %w", err)
	}

	updatedAt, err := parseESDate(doc.UpdatedAt)
	if err != nil {
		return User{}, fmt.Errorf("parsing updatedAt: %w", err)
	}

	var accountID *uuid.UUID
	if doc.AccountID != "" {
		aid, pErr := uuid.Parse(doc.AccountID)
		if pErr == nil {
			accountID = &aid
		}
	}

	return New(
		id,
		nm,
		mail.Address{Address: doc.Email},
		nil, // no password hash in ES
		dept,
		doc.Enabled,
		accountID,
		createdAt,
		updatedAt,
		nil, // deletedAt excluded by query
	), nil
}

func parseESDate(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// ES may store dates as epoch millis
		ms, pErr := strconv.ParseInt(s, 10, 64)
		if pErr != nil {
			return time.Time{}, err
		}
		return time.UnixMilli(ms), nil
	}

	return t, nil
}
