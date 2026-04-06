// Package reset_token_repo contains database-related CRUD functionality for reset tokens.
package reset_token_repo

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/reset_token"
	"github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed query/token_create.sql
	tokenCreateSQL string
	//go:embed query/token_delete.sql
	tokenDeleteSQL string
	//go:embed query/token_query_by_token.sql
	tokenQueryByTokenSQL string
)

// Store manages the set of APIs for reset token database access.
type Store struct {
	log    *logger.Service
	dbPool sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Service, dbPool *sqlx.DB) *Store {
	return &Store{
		log:    log,
		dbPool: dbPool,
	}
}

// Create inserts a new reset token into the database.
func (s *Store) Create(ctx context.Context, token reset_token.ResetToken) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, tokenCreateSQL, toTokenDB(token))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// Delete removes a reset token from the database.
func (s *Store) Delete(ctx context.Context, token reset_token.ResetToken) error {
	err := pgsql.NamedExecContext(ctx, s.log, s.dbPool, tokenDeleteSQL, toTokenDB(token))
	if err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// QueryByToken retrieves a reset token by its token string.
func (s *Store) QueryByToken(ctx context.Context, token string) (reset_token.ResetToken, error) {
	data := struct {
		Token string `db:"token"`
	}{
		Token: token,
	}

	var dbTkn tokenDB

	err := pgsql.NamedQueryStruct(ctx, s.log, s.dbPool, tokenQueryByTokenSQL, data, &dbTkn)
	if err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return reset_token.ResetToken{}, errs.Errorf(
				errs.NotFound,
				errs.CodeInvalidToken,
				"reset token not found",
			)
		}

		return reset_token.ResetToken{}, fmt.Errorf("db: %w", err)
	}

	return toTokenDomain(dbTkn), nil
}
