// Package refresh_token_service provides internal access to the domain.
package refresh_token

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/google/uuid"
)

type generator interface {
	Generate() (uuid.UUID, error)
}

type clock interface {
	Now() time.Time
}

// Service manages the set of APIs for user access.
type Service struct {
	log     logger.Logger
	storer  Storer
	uuidGen generator
	clock   clock
}

// NewService constructs a user.User internal API for use.
func NewService(
	log logger.Logger,
	storer Storer,
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

// Delete removes the specified refresh_token.
func (c *Service) Delete(ctx context.Context, tkn RefreshToken) error {
	err := c.storer.Delete(ctx, tkn)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}
