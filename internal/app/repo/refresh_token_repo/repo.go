// Package refresh_token_repo contains database-related CRUD functionality.
package refresh_token_repo

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/Housiadas/cerberus/internal/core/domain/refresh_token"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

// queries
var (
	//go:embed query/token_create.sql
	tokenCreateSql string
	//go:embed query/token_delete.sql
	tokenDeleteSql string
	//go:embed query/token_query_by_id.sql
	tokenQueryByIdSql string
)

type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

func NewStore(log *logger.Logger, db *sqlx.DB) refresh_token.Storer {
	return &Store{
		log: log,
		db:  db,
	}
}

func (s *Store) Create(ctx context.Context, token refresh_token.RefreshToken) error {
	if err := pgsql.NamedExecContext(ctx, s.log, s.db, tokenCreateSql, toTokenDB(token)); err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, token refresh_token.RefreshToken) error {
	if err := pgsql.NamedExecContext(ctx, s.log, s.db, tokenDeleteSql, toTokenDB(token)); err != nil {
		return fmt.Errorf("named_exec_context: %w", err)
	}

	return nil
}

// QueryByID gets the specified userDB from the database.
func (s *Store) QueryByID(ctx context.Context, tokenID uuid.UUID) (refresh_token.RefreshToken, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: tokenID.String(),
	}

	var dbTkn tokenDB
	if err := pgsql.NamedQueryStruct(ctx, s.log, s.db, tokenQueryByIdSql, data, &dbTkn); err != nil {
		if errors.Is(err, pgsql.ErrDBNotFound) {
			return refresh_token.RefreshToken{}, errs.Newf(errs.NotFound, "user not found")
		}
		return refresh_token.RefreshToken{}, fmt.Errorf("db: %w", err)
	}

	return toTokenDomain(dbTkn)
}
