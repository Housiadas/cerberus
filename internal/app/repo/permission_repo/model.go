package permission_repo

import (
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/google/uuid"
)

type permissionDB struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func toPermissionDB(p permission.Permission) permissionDB {
	return permissionDB{
		ID:        p.ID,
		Name:      p.Name.String(),
		CreatedAt: p.CreatedAt.UTC(),
		UpdatedAt: p.UpdatedAt.UTC(),
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
	}

	return bus, nil
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
