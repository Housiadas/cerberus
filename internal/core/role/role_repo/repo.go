// Package role_repo contains database-related CRUD functionality.
package role_repo

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/role"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/role_create.sql
	roleCreateSQL string
	//go:embed query/role_update.sql
	roleUpdateSQL string
	//go:embed query/role_delete.sql
	roleDeleteSQL string
	//go:embed query/role_query_by_id.sql
	roleQueryByIDSQL string
)

// Store manages the set of APIs for userDB database access.
type Store struct {
	log logger.Logger
	db  *sqlx.DB
}

// NewStore constructs the api for data access.
func NewStore(log logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new roleDB into the database.
func (s *Store) Create(ctx context.Context, rl role.Role) error {
	err := pgsql.NamedExecContext(ctx, s.log, pgsql.Conn(ctx, s.db), roleCreateSQL, toRoleDB(rl))
	if err != nil {
		return fmt.Errorf("error role create in db: %w", err)
	}

	return nil
}

// Update replaces a roleDB document in the database.
func (s *Store) Update(ctx context.Context, rl role.Role) error {
	err := pgsql.NamedExecContext(ctx, s.log, pgsql.Conn(ctx, s.db), roleUpdateSQL, toRoleDB(rl))
	if err != nil {
		return fmt.Errorf("error role update in db: %w", err)
	}

	return nil
}

// Delete removes a roleDB from the database.
func (s *Store) Delete(ctx context.Context, rl role.Role) error {
	err := pgsql.NamedExecContext(ctx, s.log, pgsql.Conn(ctx, s.db), roleDeleteSQL, toRoleDB(rl))
	if err != nil {
		return fmt.Errorf("error delete role in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified userDB from the database.
func (s *Store) QueryByID(ctx context.Context, roleID uuid.UUID) (role.Role, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: roleID.String(),
	}

	var dbRole roleDB

	err := pgsql.NamedQueryStruct(
		ctx,
		s.log,
		pgsql.Conn(ctx, s.db),
		roleQueryByIDSQL,
		data,
		&dbRole,
	)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return role.Role{}, fmt.Errorf("db: %w", role.ErrNotFound)
		}

		return role.Role{}, fmt.Errorf("db: %w", err)
	}

	return toRoleDomain(dbRole)
}

// Query retrieves a list of existing roles from the database.
func (s *Store) Query(
	ctx context.Context,
	filter role.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]role.Role, error) {
	col, ok := orderByFields[orderBy.Field]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrFieldNotExist, orderBy.Field)
	}

	sb := pgsql.Builder.
		Select("id", "name", "created_at", "updated_at").
		From("roles").
		Where(filterPredicates(filter)).
		OrderBy(col+" "+orderBy.Direction, "id "+orderBy.Direction).
		Limit(uint64(cur.Limit() + 1))

	if cp := cursorPredicate(cur, orderBy); cp != nil {
		sb = sb.Where(cp)
	}

	query, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var dbRoles []roleDB

	if err := pgsql.SelectSlice(ctx, s.log, pgsql.Conn(ctx, s.db), query, args, &dbRoles); err != nil {
		return nil, fmt.Errorf("select slice: %w", err)
	}

	return toRolesDomain(dbRoles)
}
