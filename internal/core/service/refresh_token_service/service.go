// Package refresh_token_service provides internal access to the domain.
package refresh_token_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/refresh_token"
	"github.com/Housiadas/cerberus/pkg/clock"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/uuidgen"
)

// Service manages the set of APIs for user access.
type Service struct {
	log     logger.Logger
	storer  refresh_token.Storer
	uuidGen uuidgen.Generator
	clock   clock.Clock
}

// New constructs a user.User internal API for use.
func New(
	log logger.Logger,
	storer refresh_token.Storer,
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

// Delete removes the specified refresh_token.
func (c *Service) Delete(ctx context.Context, tkn refresh_token.RefreshToken) error {
	err := c.storer.Delete(ctx, tkn)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
