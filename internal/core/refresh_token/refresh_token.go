package refresh_token

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token in the system.
type RefreshToken struct {
	id        uuid.UUID
	userID    uuid.UUID
	token     string
	expiresAt time.Time
	createdAt time.Time
	revoked   bool
}

// New constructs a RefreshToken from its individual fields.
func New(
	id uuid.UUID,
	userID uuid.UUID,
	token string,
	expiresAt time.Time,
	createdAt time.Time,
	revoked bool,
) RefreshToken {
	return RefreshToken{
		id:        id,
		userID:    userID,
		token:     token,
		expiresAt: expiresAt,
		createdAt: createdAt,
		revoked:   revoked,
	}
}

// ID returns the refresh token ID.
func (rt RefreshToken) ID() uuid.UUID { return rt.id }

// UserID returns the user ID associated with this token.
func (rt RefreshToken) UserID() uuid.UUID { return rt.userID }

// Token returns the token string.
func (rt RefreshToken) Token() string { return rt.token }

// ExpiresAt returns the expiration time.
func (rt RefreshToken) ExpiresAt() time.Time { return rt.expiresAt }

// CreatedAt returns the creation time.
func (rt RefreshToken) CreatedAt() time.Time { return rt.createdAt }

// Revoked returns whether the token has been revoked.
func (rt RefreshToken) Revoked() bool { return rt.revoked }

// WithRevoked returns a new RefreshToken with the given revoked state.
func (rt RefreshToken) WithRevoked(revoked bool) RefreshToken {
	rt.revoked = revoked

	return rt
}
