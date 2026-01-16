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
	UserID    string `json:"userId"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
	CreatedAt string `json:"createdAt"`
	Revoked   bool   `json:"revoked"`
}

// Encode implements the encoder interface.
func (r RefreshToken) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, web.ContentTypeJSON, fmt.Errorf("refresh token encode error: %w", err)
	}

	return data, web.ContentTypeJSON, nil
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
