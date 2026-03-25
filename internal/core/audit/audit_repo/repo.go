// Package audit_repo contains auditDB-related CRUD functionality.
package audit_repo

import (
	"bytes"
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

// queries.
var (
	//go:embed query/audit_create.sql
	auditCreateSQL string
	//go:embed query/audit_query.sql
	auditQuerySQL string
)

// Store manages the set of APIs for auditDB database access.
type Store struct {
	log    *logger.Service
	dbPool sqlx.ExtContext
}

// NewStore constructs the API for data access.
func NewStore(
	log *logger.Service,
	dbPool *sqlx.DB,
) *Store {
	return &Store{
		log:    log,
		dbPool: dbPool,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB with a sqlx DB that is
// running within a transaction.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (audit.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("audit transaction init error: %w", err)
	}

	store := Store{
		log:    s.log,
		dbPool: ec,
	}

	return &store, nil
}

// Create inserts a new auditDB record into the database.
func (s *Store) Create(ctx context.Context, a audit.Audit) error {
	dbAudit := toDBAudit(a)

	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, auditCreateSQL, dbAudit)
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
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(auditQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbAudits []auditDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.dbPool, buf.String(), data, &dbAudits)
	if err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusAudits(dbAudits)
}
