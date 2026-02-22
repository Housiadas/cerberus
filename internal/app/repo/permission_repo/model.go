package permission_repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/google/uuid"
)

type permissionDB struct {
	ID        uuid.UUID    `db:"id"`
	Name      string       `db:"name"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func toPermissionDB(p permission.Permission) permissionDB {
	return permissionDB{
		ID:        p.ID,
		Name:      p.Name.String(),
		CreatedAt: p.CreatedAt.UTC(),
		UpdatedAt: p.UpdatedAt.UTC(),
		DeletedAt: toNullTime(p.DeletedAt),
	}
}

func toPermissionDomain(db permissionDB) (permission.Permission, error) {
	nme, err := name.Parse(db.Name)
	if err != nil {
		return permission.Permission{}, fmt.Errorf("parse name: %w", err)
	}

	bus := permission.Permission{
		ID:        db.ID,
		Name:      nme,
		CreatedAt: db.CreatedAt.In(time.UTC),
		UpdatedAt: db.UpdatedAt.In(time.UTC),
		DeletedAt: fromNullTime(db.DeletedAt),
	}

	return bus, nil
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

func toPermissionsDomain(dbs []permissionDB) ([]permission.Permission, error) {
	bus := make([]permission.Permission, len(dbs))

	for i, db := range dbs {
		var err error

		bus[i], err = toPermissionDomain(db)
		if err != nil {
			return nil, fmt.Errorf("to permission object error: %w", err)
		}
	}

	return bus, nil
}
