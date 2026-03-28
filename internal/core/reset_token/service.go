// Package reset_token_service provides access to the reset token domain.
package reset_token

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const resetTokenTTL = 15 * time.Minute

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}

// Service manages the set of APIs for reset token access.
type Service struct {
	storer  Storer
	uuidGen generator
	clock   clock
}

// NewService constructs the service.
func NewService(
	storer Storer,
	uuidGen generator,
	clock clock,
) *Service {
	return &Service{
		storer:  storer,
		uuidGen: uuidGen,
		clock:   clock,
	}
}

// Create generates and persists a new password reset token for the given user.
func (s *Service) Create(ctx context.Context, userID uuid.UUID) (ResetToken, error) {
	id, err := s.uuidGen.Generate()
	if err != nil {
		return ResetToken{}, fmt.Errorf("uuid: %w", err)
	}

	tokenID, err := s.uuidGen.Generate()
	if err != nil {
		return ResetToken{}, fmt.Errorf("uuid: %w", err)
	}

	now := s.clock.Now()
	tkn := New(
		id,
		userID,
		tokenID.String(),
		now.UTC().Add(resetTokenTTL),
		now,
		false,
	)

	err = s.storer.Create(ctx, tkn)
	if err != nil {
		return ResetToken{}, fmt.Errorf("create: %w", err)
	}

	return tkn, nil
}

// QueryByToken retrieves a reset token by its token string.
func (s *Service) QueryByToken(ctx context.Context, token string) (ResetToken, error) {
	tkn, err := s.storer.QueryByToken(ctx, token)
	if err != nil {
		return ResetToken{}, fmt.Errorf("query by token: %w", err)
	}

	return tkn, nil
}

// Delete removes the specified reset token.
func (s *Service) Delete(ctx context.Context, tkn ResetToken) error {
	err := s.storer.Delete(ctx, tkn)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
