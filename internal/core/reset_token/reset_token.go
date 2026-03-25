// Package reset_token defines the password reset token domain model.
package reset_token

import (
	"time"

	"github.com/google/uuid"
)

// ResetToken represents a short-lived password reset token.
type ResetToken struct {
	id        uuid.UUID
	userID    uuid.UUID
	token     string
	expiresAt time.Time
	createdAt time.Time
	used      bool
}

// New constructs a ResetToken from its individual fields.
func New(
	id uuid.UUID,
	userID uuid.UUID,
	token string,
	expiresAt time.Time,
	createdAt time.Time,
	used bool,
) ResetToken {
	return ResetToken{
		id:        id,
		userID:    userID,
		token:     token,
		expiresAt: expiresAt,
		createdAt: createdAt,
		used:      used,
	}
}

// ID returns the reset token ID.
func (rt ResetToken) ID() uuid.UUID { return rt.id }

// UserID returns the user ID associated with this token.
func (rt ResetToken) UserID() uuid.UUID { return rt.userID }

// Token returns the token string.
func (rt ResetToken) Token() string { return rt.token }

// ExpiresAt returns the expiration time.
func (rt ResetToken) ExpiresAt() time.Time { return rt.expiresAt }

// CreatedAt returns the creation time.
func (rt ResetToken) CreatedAt() time.Time { return rt.createdAt }

// Used returns whether the token has been used.
func (rt ResetToken) Used() bool { return rt.used }

// IsExpired reports whether the token is past its expiration time.
func (rt ResetToken) IsExpired() bool {
	return time.Now().UTC().After(rt.expiresAt)
}
