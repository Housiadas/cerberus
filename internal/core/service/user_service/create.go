package user_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
)

// Create adds a new User to the system.
func (c *Service) Create(ctx context.Context, nu user.NewUser) (user.User, error) {
	id, err := c.uuidGen.Generate()
	if err != nil {
		return user.User{}, fmt.Errorf("uuid v7 error: %w", err)
	}

	hash, err := c.hasher.Hash(nu.Password.String())
	if err != nil {
		return user.User{}, fmt.Errorf("generate_from_password: %w", err)
	}

	now := c.clock.Now()
	usr := user.User{
		ID:           id,
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: hash,
		Department:   nu.Department,
		Enabled:      true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = c.storer.Create(ctx, usr)
	if err != nil {
		return user.User{}, fmt.Errorf("create: %w", err)
	}

	return usr, nil
}
