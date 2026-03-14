package user_repo

import (
	"database/sql"
	"fmt"
	"net/mail"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/google/uuid"
)

type userDB struct {
	ID           uuid.UUID      `db:"id"`
	Name         string         `db:"name"`
	Email        string         `db:"email"`
	PasswordHash []byte         `db:"password_hash"`
	Department   sql.NullString `db:"department"`
	AccountID    uuid.NullUUID  `db:"account_id"`
	Enabled      bool           `db:"enabled"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	DeletedAt    sql.NullTime   `db:"deleted_at"`
}

func toUserDB(usr user.User) userDB {
	return userDB{
		ID:           usr.ID(),
		Name:         usr.Name().String(),
		Email:        usr.Email().Address,
		PasswordHash: usr.PasswordHash(),
		Department: sql.NullString{
			String: usr.Department().String(),
			Valid:  usr.Department().Valid(),
		},
		AccountID: toNullUUID(usr.AccountID()),
		Enabled:   usr.Enabled(),
		CreatedAt: usr.CreatedAt().UTC(),
		UpdatedAt: usr.UpdatedAt().UTC(),
		DeletedAt: toNullTime(usr.DeletedAt()),
	}
}

func toUserDomain(db userDB) (user.User, error) {
	addr := mail.Address{
		Address: db.Email,
	}

	nme, err := name.Parse(db.Name)
	if err != nil {
		return user.User{}, fmt.Errorf("parse name: %w", err)
	}

	department, err := name.ParseNull(db.Department.String)
	if err != nil {
		return user.User{}, fmt.Errorf("parse department: %w", err)
	}

	return user.New(
		db.ID,
		nme,
		addr,
		db.PasswordHash,
		department,
		fromNullUUID(db.AccountID),
		db.Enabled,
		db.CreatedAt.In(time.UTC),
		db.UpdatedAt.In(time.UTC),
		fromNullTime(db.DeletedAt),
	), nil
}

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}

	return uuid.NullUUID{UUID: *id, Valid: true}
}

func fromNullUUID(nu uuid.NullUUID) *uuid.UUID {
	if !nu.Valid {
		return nil
	}

	return &nu.UUID
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

func toUsersDomain(dbs []userDB) ([]user.User, error) {
	bus := make([]user.User, len(dbs))

	for i, db := range dbs {
		var err error

		bus[i], err = toUserDomain(db)
		if err != nil {
			return nil, err
		}
	}

	return bus, nil
}
