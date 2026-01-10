package refresh_token

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token in the system.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
	Revoked   bool
}
