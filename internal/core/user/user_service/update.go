package user_service

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/user"
)

// Update modifies information about a user.User.
func (c *Service) Update(
	ctx context.Context,
	usr user.User,
	uu user.UpdateUser,
) (user.User, error) {
	if uu.Name != nil {
		usr = usr.WithName(*uu.Name)
	}

	if uu.Email != nil {
		usr = usr.WithEmail(*uu.Email)
	}

	if uu.Password != nil {
		pw, err := c.hasher.Hash(uu.Password.String())
		if err != nil {
			return user.User{}, fmt.Errorf("generate_from_password: %w", err)
		}

		usr = usr.WithPasswordHash(pw)
	}

	if uu.Department != nil {
		usr = usr.WithDepartment(*uu.Department)
	}

	if uu.Enabled != nil {
		usr = usr.WithEnabled(*uu.Enabled)
	}

	usr = usr.WithUpdatedAt(c.clock.Now())

	err := c.storer.Update(ctx, usr)
	if err != nil {
		return user.User{}, fmt.Errorf("update: %w", err)
	}

	return usr, nil
}
