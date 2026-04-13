package reset_token

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const resetTokenTTL = 15 * time.Minute

// Service manages the set of APIs for reset token access.
type Service struct {
	storer  storer
	uuidGen generator
	clock   clock
}

// NewService constructs the service.
func NewService(
	storer storer,
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

	params := toCreateResetTokenParams(tkn)

	dbToken, err := s.storer.CreateResetToken(ctx, params)
	if err != nil {
		return ResetToken{}, fmt.Errorf("create: %w", err)
	}

	return toDomainResetToken(dbToken), nil
}

// QueryByToken retrieves a reset token by its token string.
func (s *Service) QueryByToken(ctx context.Context, token string) (ResetToken, error) {
	dbToken, err := s.storer.GetResetTokenByToken(ctx, token)
	if err != nil {
		return ResetToken{}, fmt.Errorf("query by token: %w", err)
	}

	return toDomainResetToken(dbToken), nil
}

// Delete removes the specified reset token.
func (s *Service) Delete(ctx context.Context, tkn ResetToken) error {
	err := s.storer.DeleteResetToken(ctx, tkn.ID())
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
