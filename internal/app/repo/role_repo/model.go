package role_repo

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/core/domain/name"
	"github.com/Housiadas/cerberus/internal/core/domain/role"
)

type roleDB struct {
	ID          uuid.UUID `db:"user_id"`
	Name        string    `db:"name"`
	DateCreated time.Time `db:"date_created"`
	DateUpdated time.Time `db:"date_updated"`
}

func toRoleDB(rl role.Role) roleDB {
	return roleDB{
		ID:          rl.ID,
		Name:        rl.Name.String(),
		DateCreated: rl.DateCreated.UTC(),
		DateUpdated: rl.DateUpdated.UTC(),
	}
}

func toRoleDomain(db roleDB) (role.Role, error) {
	nme, err := name.Parse(db.Name)
	if err != nil {
		return role.Role{}, fmt.Errorf("parse name: %w", err)
	}

	bus := role.Role{
		ID:          db.ID,
		Name:        nme,
		DateCreated: db.DateCreated.In(time.Local),
		DateUpdated: db.DateUpdated.In(time.Local),
	}

	return bus, nil
}

func toRolesDomain(dbs []roleDB) ([]role.Role, error) {
	bus := make([]role.Role, len(dbs))

	for i, db := range dbs {
		var err error
		bus[i], err = toRoleDomain(db)
		if err != nil {
			return nil, err
		}
	}

	return bus, nil
}
