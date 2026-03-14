// Package plan_repo contains database-related CRUD functionality.
package plan_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/plan"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/plan_create.sql
	planCreateSQL string
	//go:embed query/plan_update.sql
	planUpdateSQL string
	//go:embed query/plan_delete.sql
	planDeleteSQL string
	//go:embed query/plan_query.sql
	planQuerySQL string
	//go:embed query/plan_query_by_id.sql
	planQueryByIDSQL string
)

// Store manages the set of APIs for plan database access.
type Store struct {
	log logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (plan.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("plan transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new plan into the database.
func (s *Store) Create(ctx context.Context, p plan.Plan) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, planCreateSQL, toPlanDB(p))
	if err != nil {
		return fmt.Errorf("error plan create in db: %w", err)
	}

	return nil
}

// Update replaces a plan document in the database.
func (s *Store) Update(ctx context.Context, p plan.Plan) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, planUpdateSQL, toPlanDB(p))
	if err != nil {
		return fmt.Errorf("error plan update in db: %w", err)
	}

	return nil
}

// Delete removes a plan from the database.
func (s *Store) Delete(ctx context.Context, p plan.Plan) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, planDeleteSQL, toPlanDB(p))
	if err != nil {
		return fmt.Errorf("error delete plan in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified plan from the database.
func (s *Store) QueryByID(ctx context.Context, planID uuid.UUID) (plan.Plan, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: planID.String(),
	}

	var dbPlan planDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, planQueryByIDSQL, data, &dbPlan)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return plan.Plan{}, fmt.Errorf("db: %w", plan.ErrNotFound)
		}

		return plan.Plan{}, fmt.Errorf("db: %w", err)
	}

	return toPlanDomain(dbPlan)
}

// Query retrieves a list of existing plans from the database.
func (s *Store) Query(
	ctx context.Context,
	filter plan.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]plan.Plan, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(planQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbPlans []planDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbPlans)
	if err != nil {
		return nil, fmt.Errorf("error query plan in db: %w", err)
	}

	return toPlansDomain(dbPlans)
}
