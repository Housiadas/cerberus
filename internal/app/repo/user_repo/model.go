package user_repo

import (
	"database/sql"
	"fmt"
	"net/mail"
	"time"

	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
)

type userDB struct {
	ID           uuid.UUID      `db:"id"`
	Name         string         `db:"name"`
	Email        string         `db:"email"`
	PasswordHash []byte         `db:"password_hash"`
	Department   sql.NullString `db:"department"`
	Enabled      bool           `db:"enabled"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
}

func toUserDB(usr user.User) userDB {
	return userDB{
		ID:           usr.ID,
		Name:         usr.Name.String(),
		Email:        usr.Email.Address,
		PasswordHash: usr.PasswordHash,
		Department: sql.NullString{
			String: usr.Department.String(),
			Valid:  usr.Department.Valid(),
		},
		Enabled:   usr.Enabled,
		CreatedAt: usr.CreatedAt.UTC(),
		UpdatedAt: usr.UpdatedAt.UTC(),
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

	bus := user.User{
		ID:           db.ID,
		Name:         nme,
		Email:        addr,
		PasswordHash: db.PasswordHash,
		Enabled:      db.Enabled,
		Department:   department,
		CreatedAt:    db.CreatedAt.In(time.Local),
		UpdatedAt:    db.UpdatedAt.In(time.Local),
	}

	return bus, nil
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
