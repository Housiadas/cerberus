package user_service

import (
	"github.com/Housiadas/cerberus/internal/core/domain/user"
)

// VerifyPassword checks whether the given plaintext password matches the
// user's stored password hash. Returns user.ErrAuthenticationFailure on mismatch.
func (c *Service) VerifyPassword(usr user.User, plainPassword string) error {
	err := c.hasher.Compare(usr.PasswordHash(), plainPassword)
	if err != nil {
		return user.ErrAuthenticationFailure
	}

	return nil
}
