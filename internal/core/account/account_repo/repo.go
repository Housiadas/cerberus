// Package account_repo contains database related CRUD functionality.
package account_repo

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/account"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/account_create.sql
	accountCreateSQL string
	//go:embed query/account_update.sql
	accountUpdateSQL string
	//go:embed query/account_delete.sql
	accountDeleteSQL string
	//go:embed query/account_query_by_id.sql
	accountQueryByIDSQL string
)

// Store manages the set of APIs for account database access.
type Store struct {
	log logger.Logger
	db  *sqlx.DB
}

// NewStore constructs the api for data access.
func NewStore(
	log logger.Logger,
	db *sqlx.DB,
) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Query retrieves a list of existing accounts from the database.
func (s *Store) Query(
	ctx context.Context,
	filter account.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]account.Account, error) {
	col, ok := orderByFields[orderBy.Field]
	if !ok {
		return nil, fmt.Errorf("%w: %s", errOrderFieldNotFound, orderBy.Field)
	}

	sb := pgsql.Builder.
		Select("id", "name", "stripe_customer_id", "created_at", "updated_at").
		From("accounts").
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

	var dbAccs []accountDB

	err = pgsql.SelectSlice(ctx, s.log, pgsql.Conn(ctx, s.db), query, args, &dbAccs)
	if err != nil {
		return nil, fmt.Errorf("select slice: %w", err)
	}

	return toAccountsDomain(dbAccs), nil
}
