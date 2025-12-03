package permission_repo

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
)

type permissionDB struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

func toPermissionDB(p permission.Permission) permissionDB {
	return permissionDB{
		ID:          p.ID,
		Name:        p.Name.String(),
		DateCreated: p.DateCreated.UTC(),
		DateUpdated: p.DateUpdated.UTC(),
	}
}

func toPermissionDomain(db permissionDB) (permission.Permission, error) {
	nme, err := name.Parse(db.Name)
	if err != nil {
		return permission.Permission{}, fmt.Errorf("parse name: %w", err)
	}

	bus := permission.Permission{
		ID:          db.ID,
		Name:        nme,
		DateCreated: db.DateCreated.In(time.Local),
		DateUpdated: db.DateUpdated.In(time.Local),
	}

	return bus, nil
}

func toPermissionsDomain(dbs []permissionDB) ([]permission.Permission, error) {
	bus := make([]permission.Permission, len(dbs))

	for i, db := range dbs {
		var err error
		bus[i], err = toPermissionDomain(db)
		if err != nil {
			return nil, err
		}
	}

	return bus, nil
}
