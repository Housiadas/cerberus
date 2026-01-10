package user_service

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/Housiadas/cerberus/internal/core/domain/user"
)

// Authenticate finds a user by their email and verifies their password. On
// success, it returns a Claims User representing this user. The claims can be
// used to generate a token for future authentication.
func (c *Service) Authenticate(
	ctx context.Context,
	email mail.Address,
	password string,
) (user.User, error) {
	usr, err := c.QueryByEmail(ctx, email)
	if err != nil {
		return user.User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	if err := c.hasher.Compare(usr.PasswordHash, password); err != nil {
		return user.User{}, fmt.Errorf(
			"compare_hash_and_password: %w",
			user.ErrAuthenticationFailure,
		)
	}

	return usr, nil
}
