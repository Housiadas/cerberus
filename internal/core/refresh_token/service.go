// Package refresh_token provides internal access to the domain.
package refresh_token

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/google/uuid"
)

// Service manages the set of APIs for user access.
type Service struct {
	log     logger.Logger
	storer  storer
	uuidGen generator
	clock   clock
}

// NewService constructs a user.User internal API for use.
func NewService(
	log logger.Logger,
	storer storer,
	uuidGen generator,
	clock clock,
) *Service {
	return &Service{
		log:     log,
		storer:  storer,
		uuidGen: uuidGen,
		clock:   clock,
	}
}

// Create adds a new refresh token to the system.
func (c *Service) Create(
	ctx context.Context,
	userID uuid.UUID,
	refreshTokenTTL time.Duration,
) (RefreshToken, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return RefreshToken{}, fmt.Errorf("uuid: %w", err)
	}

	tokenID, err := c.uuidGen.Generate()
	if err != nil {
		return RefreshToken{}, fmt.Errorf("uuid: %w", err)
	}

	now := c.clock.Now()
	tkn := New(
		id,
		userID,
		tokenID.String(),
		now.UTC().Add(refreshTokenTTL),
		now,
		false,
	)

	params := toCreateRefreshTokenParams(tkn)

	dbToken, err := c.storer.CreateRefreshToken(ctx, params)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("create: %w", err)
	}

	return toDomainRefreshToken(dbToken), nil
}

func (c *Service) QueryByToken(
	ctx context.Context,
	token string,
) (RefreshToken, error) {
	dbToken, err := c.storer.GetRefreshTokenByToken(ctx, token)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("query by token: %w", err)
	}

	return toDomainRefreshToken(dbToken), nil
}

// Revoke revokes the specified refresh_token.
func (c *Service) Revoke(ctx context.Context, tkn RefreshToken) error {
	err := c.storer.RevokeRefreshToken(ctx, tkn.Token())
	if err != nil {
		return fmt.Errorf("revoke: %w", err)
	}

	return nil
}
