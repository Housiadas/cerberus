package user_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
)

// Update modifies information about a user.User.
func (c *Service) Update(ctx context.Context, usr user.User, uu user.UpdateUser) (user.User, error) {
	if uu.Name != nil {
		usr.Name = *uu.Name
	}

	if uu.Email != nil {
		usr.Email = *uu.Email
	}

	if uu.Password != nil {
		pw, err := c.hasher.Hash(uu.Password.String())
		if err != nil {
			return user.User{}, fmt.Errorf("generate_from_password: %w", err)
		}
		usr.PasswordHash = pw
	}

	if uu.Department != nil {
		usr.Department = *uu.Department
	}

	if uu.Enabled != nil {
		usr.Enabled = *uu.Enabled
	}
	usr.UpdatedAt = c.clock.Now()

	if err := c.storer.Update(ctx, usr); err != nil {
		return user.User{}, fmt.Errorf("update: %w", err)
	}

	return usr, nil
}
