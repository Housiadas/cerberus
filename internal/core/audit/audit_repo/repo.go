// Package audit_repo contains auditDB-related CRUD functionality.
package audit_repo

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/jmoiron/sqlx"
)

//go:embed query/audit_create.sql
var auditCreateSQL string

// Store manages the set of APIs for auditDB database access.
type Store struct {
	log logger.Logger
	db  *sqlx.DB
}

// NewStore constructs the API for data access.
func NewStore(
	log logger.Logger,
	db *sqlx.DB,
) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new auditDB record into the database.
func (s *Store) Create(ctx context.Context, a audit.Audit) error {
	dbAudit := toDBAudit(a)

	err := pgsql.NamedExecContext(ctx, s.log, pgsql.Conn(ctx, s.db), auditCreateSQL, dbAudit)
	if err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of existing audit records from the database.
func (s *Store) Query(
	ctx context.Context,
	filter audit.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]audit.Audit, error) {
	col, ok := orderByFields[orderBy.Field]
	if !ok {
		return nil, fmt.Errorf("%w: %s", errOrderFieldNotFound, orderBy.Field)
	}

	sb := pgsql.Builder.
		Select(
			"id", "obj_id", "obj_entity", "obj_name",
			"actor_id", "action", "data", "message", "timestamp",
		).
		From("audit").
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

	var dbAudits []auditDB

	err = pgsql.SelectSlice(ctx, s.log, pgsql.Conn(ctx, s.db), query, args, &dbAudits)
	if err != nil {
		return nil, fmt.Errorf("select slice: %w", err)
	}

	return toBusAudits(dbAudits)
}
