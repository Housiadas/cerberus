package role_repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/role"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/google/uuid"
)

type roleDB struct {
	ID        uuid.UUID    `db:"id"`
	Name      string       `db:"name"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func toRoleDB(rl role.Role) roleDB {
	return roleDB{
		ID:        rl.ID(),
		Name:      rl.Name().String(),
		CreatedAt: rl.CreatedAt().UTC(),
		UpdatedAt: rl.UpdatedAt().UTC(),
		DeletedAt: toNullTime(rl.DeletedAt()),
	}
}

func toRoleDomain(db roleDB) (role.Role, error) {
	nme, err := name.Parse(db.Name)
	if err != nil {
		return role.Role{}, fmt.Errorf("parse name: %w", err)
	}

	return role.New(
		db.ID,
		nme,
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

func toRolesDomain(dbs []roleDB) ([]role.Role, error) {
	bus := make([]role.Role, len(dbs))

	for i, db := range dbs {
		var err error

		bus[i], err = toRoleDomain(db)
		if err != nil {
			return nil, fmt.Errorf("to roles domain error: %w", err)
		}
	}

	return bus, nil
}
