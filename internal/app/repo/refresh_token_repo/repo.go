// Package refresh_token_repo contains database-related CRUD functionality.
package refresh_token_repo

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/refresh_token"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/web/errs"
	"github.com/jmoiron/sqlx"
)

// queries.
var (
	//go:embed query/token_create.sql
	tokenCreateSQL string
	//go:embed query/token_delete.sql
	tokenDeleteSQL string
	//go:embed query/token_revoke.sql
	tokenRevokeSQL string
	//go:embed query/token_query_by_token.sql
	tokenQueryByTokenSQL string
)

type Store struct {
	log    logger.Logger
	dbPool sqlx.ExtContext
}

func NewStore(log logger.Logger, dbPool *sqlx.DB) *Store {
	return &Store{
		log:    log,
		dbPool: dbPool,
	}
}

func (s *Store) Create(ctx context.Context, token refresh_token.RefreshToken) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, tokenCreateSQL, toTokenDB(token))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, token refresh_token.RefreshToken) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, tokenDeleteSQL, toTokenDB(token))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

func (s *Store) Revoke(ctx context.Context, token refresh_token.RefreshToken) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, tokenRevokeSQL, toTokenDB(token))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

func (s *Store) QueryByToken(
	ctx context.Context,
	token string,
) (refresh_token.RefreshToken, error) {
	data := struct {
		Token string `db:"token"`
	}{
		Token: token,
	}

	var dbTkn tokenDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.dbPool, tokenQueryByTokenSQL, data, &dbTkn)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return refresh_token.RefreshToken{}, errs.Errorf(
				errs.NotFound,
				"refresh token not found",
			)
		}

		return refresh_token.RefreshToken{}, fmt.Errorf("db: %w", err)
	}

	return toTokenDomain(dbTkn)
}
