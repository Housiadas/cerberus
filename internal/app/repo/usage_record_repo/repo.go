// Package usage_record_repo contains database-related CRUD functionality.
package usage_record_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/usage_record"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/usage_record_create.sql
	usageRecordCreateSQL string
	//go:embed query/usage_record_query.sql
	usageRecordQuerySQL string
	//go:embed query/usage_record_query_by_id.sql
	usageRecordQueryByIDSQL string
)

// Store manages the set of APIs for usage_record database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (usage_record.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("usage record transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new usage record into the database.
func (s *Store) Create(ctx context.Context, ur usage_record.UsageRecord) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, usageRecordCreateSQL, toUsageRecordDB(ur))
	if err != nil {
		return fmt.Errorf("error usage record create in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified usage record from the database.
func (s *Store) QueryByID(ctx context.Context, usageRecordID uuid.UUID) (usage_record.UsageRecord, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: usageRecordID.String(),
	}

	var dbUR usageRecordDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, usageRecordQueryByIDSQL, data, &dbUR)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return usage_record.UsageRecord{}, fmt.Errorf("db: %w", usage_record.ErrNotFound)
		}

		return usage_record.UsageRecord{}, fmt.Errorf("db: %w", err)
	}

	return toUsageRecordDomain(dbUR), nil
}

// Query retrieves a list of existing usage records from the database.
func (s *Store) Query(
	ctx context.Context,
	filter usage_record.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]usage_record.UsageRecord, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(usageRecordQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbURs []usageRecordDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbURs)
	if err != nil {
		return nil, fmt.Errorf("error query usage record in db: %w", err)
	}

	return toUsageRecordsDomain(dbURs), nil
}
