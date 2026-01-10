package refresh_token_usecase

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/domain/refresh_token"
	"github.com/Housiadas/cerberus/pkg/web/errs"
	"github.com/google/uuid"
)

// =============================================================================

// RefreshToken represents information about an individual user.
type RefreshToken struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
	Revoked   bool   `json:"revoked"`
}

// Encode implements the encoder interface.
func (r RefreshToken) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)

	return data, "application/json", fmt.Errorf("refresh token error: %w", err)
}

func toAppToken(r refresh_token.RefreshToken) RefreshToken {
	return RefreshToken{
		ID:        r.ID.String(),
		UserID:    r.UserID.String(),
		Token:     r.Token,
		ExpiresAt: r.ExpiresAt.Format(time.RFC3339),
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
		Revoked:   r.Revoked,
	}
}

func toAppTokens(tkns []refresh_token.RefreshToken) []RefreshToken {
	appRoles := make([]RefreshToken, len(tkns))
	for i, rl := range tkns {
		appRoles[i] = toAppToken(rl)
	}

	return appRoles
}

func toCoreToken(r RefreshToken) (refresh_token.RefreshToken, error) {
	var errors errs.FieldErrors

	id, err := uuid.Parse(r.ID)
	if err != nil {
		errors.Add("id", err)
	}

	userID, err := uuid.Parse(r.UserID)
	if err != nil {
		errors.Add("user_id", err)
	}

	expiresAt, err := time.Parse(time.RFC3339, r.ExpiresAt)
	if err != nil {
		errors.Add("expires_at", err)
	}

	createdAt, err := time.Parse(time.RFC3339, r.CreatedAt)
	if err != nil {
		errors.Add("created_at", err)
	}

	if len(errors) > 0 {
		return refresh_token.RefreshToken{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return refresh_token.RefreshToken{
		ID:        id,
		UserID:    userID,
		Token:     r.Token,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
		Revoked:   r.Revoked,
	}, nil
}
