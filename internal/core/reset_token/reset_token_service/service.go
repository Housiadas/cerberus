// Package reset_token_service provides access to the reset token domain.
package reset_token_service

import (
	"context"
	"fmt"
	"time"

	reset_token2 "github.com/Housiadas/cerberus/internal/core/reset_token"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
	"github.com/google/uuid"
)

const resetTokenTTL = 15 * time.Minute

// Service manages the set of APIs for reset token access.
type Service struct {
	log     logger.Logger
	storer  reset_token2.Storer
	uuidGen uuidgen.Generator
	clock   clock.Clock
}

// New constructs the service.
func New(
	log logger.Logger,
	storer reset_token2.Storer,
	uuidGen uuidgen.Generator,
	clock clock.Clock,
) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
		clock:   clock,
	}
}

// Create generates and persists a new password reset token for the given user.
func (s *Service) Create(ctx context.Context, userID uuid.UUID) (reset_token2.ResetToken, error) {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return reset_token2.ResetToken{}, fmt.Errorf("uuid: %w", err)
	}

	tokenID, err := s.uuidGen.Generate()
	if err != nil {
		return reset_token2.ResetToken{}, fmt.Errorf("uuid: %w", err)
	}

	now := s.clock.Now()
	tkn := reset_token2.New(
		id,
		userID,
		tokenID.String(),
		now.UTC().Add(resetTokenTTL),
		now,
		false,
	)

	err = s.storer.Create(ctx, tkn)
	if err != nil {
		return reset_token2.ResetToken{}, fmt.Errorf("create: %w", err)
	}

	return tkn, nil
}

// QueryByToken retrieves a reset token by its token string.
func (s *Service) QueryByToken(ctx context.Context, token string) (reset_token2.ResetToken, error) {
	tkn, err := s.storer.QueryByToken(ctx, token)
	if err != nil {
		return reset_token2.ResetToken{}, fmt.Errorf("query by token: %w", err)
	}

	return tkn, nil
}

// Delete removes the specified reset token.
func (s *Service) Delete(ctx context.Context, tkn reset_token2.ResetToken) error {
	err := s.storer.Delete(ctx, tkn)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
