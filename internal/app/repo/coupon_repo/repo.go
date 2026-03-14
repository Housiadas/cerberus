// Package coupon_repo contains database-related CRUD functionality.
package coupon_repo

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/coupon"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/coupon_create.sql
	couponCreateSQL string
	//go:embed query/coupon_update.sql
	couponUpdateSQL string
	//go:embed query/coupon_delete.sql
	couponDeleteSQL string
	//go:embed query/coupon_query.sql
	couponQuerySQL string
	//go:embed query/coupon_query_by_id.sql
	couponQueryByIDSQL string
	//go:embed query/coupon_query_by_code.sql
	couponQueryByCodeSQL string
)

// Store manages the set of APIs for coupon database access.
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
func (s *Store) NewWithTx(tx pgsql.CommitRollbacker) (coupon.Storer, error) {
	ec, err := pgsql.GetExtContext(tx)
	if err != nil {
		return nil, fmt.Errorf("coupon transaction init error: %w", err)
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new coupon into the database.
func (s *Store) Create(ctx context.Context, cpn coupon.Coupon) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, couponCreateSQL, toCouponDB(cpn))
	if err != nil {
		return fmt.Errorf("error coupon create in db: %w", err)
	}

	return nil
}

// Update replaces a coupon document in the database.
func (s *Store) Update(ctx context.Context, cpn coupon.Coupon) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, couponUpdateSQL, toCouponDB(cpn))
	if err != nil {
		return fmt.Errorf("error coupon update in db: %w", err)
	}

	return nil
}

// Delete removes a coupon from the database.
func (s *Store) Delete(ctx context.Context, cpn coupon.Coupon) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.db, couponDeleteSQL, toCouponDB(cpn))
	if err != nil {
		return fmt.Errorf("error delete coupon in db: %w", err)
	}

	return nil
}

// QueryByID gets the specified coupon from the database.
func (s *Store) QueryByID(ctx context.Context, couponID uuid.UUID) (coupon.Coupon, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: couponID.String(),
	}

	var dbCpn couponDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, couponQueryByIDSQL, data, &dbCpn)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return coupon.Coupon{}, fmt.Errorf("db: %w", coupon.ErrNotFound)
		}

		return coupon.Coupon{}, fmt.Errorf("db: %w", err)
	}

	return toCouponDomain(dbCpn)
}

// QueryByCode gets the specified coupon by code from the database.
func (s *Store) QueryByCode(ctx context.Context, code string) (coupon.Coupon, error) {
	data := struct {
		Code string `db:"code"`
	}{
		Code: code,
	}

	var dbCpn couponDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.db, couponQueryByCodeSQL, data, &dbCpn)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return coupon.Coupon{}, fmt.Errorf("db: %w", coupon.ErrNotFound)
		}

		return coupon.Coupon{}, fmt.Errorf("db: %w", err)
	}

	return toCouponDomain(dbCpn)
}

// Query retrieves a list of existing coupons from the database.
func (s *Store) Query(
	ctx context.Context,
	filter coupon.QueryFilter,
	orderBy order.By,
	cur cursor.Cursor,
) ([]coupon.Coupon, error) {
	data := map[string]any{
		"limit": cur.Limit() + 1,
	}

	buf := bytes.NewBufferString(couponQuerySQL)
	applyFilter(filter, data, buf)
	applyCursor(cur, orderBy, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" FETCH NEXT :limit ROWS ONLY")

	var dbCpns []couponDB

	err = pgsql.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbCpns)
	if err != nil {
		return nil, fmt.Errorf("error query coupon in db: %w", err)
	}

	return toCouponsDomain(dbCpns)
}
