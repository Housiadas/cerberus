package auth_usecase

import (
	"encoding/json"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/pkg/errs"
)

// AuthLogin defines the data needed to authenticate a user.
type AuthLogin struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password"`
}

// Encode implements the encoder interface.
func (l *AuthLogin) Encode() ([]byte, string, error) {
	data, err := json.Marshal(l)
	return data, "application/json", err
}

// Validate checks the data in the model is considered clean.
func (l *AuthLogin) Validate() error {
	if err := validation.Check(l); err != nil {
		return errs.Newf(errs.InvalidArgument, "validation: %s", err)
	}

	return nil
}

// Decode implements the decoder interface.
func (l *AuthLogin) Decode(data []byte) error {
	return json.Unmarshal(data, l)
}

// =================================================================

type AuthRefreshToken struct {
	Token string `json:"refresh_token"`
}

// Encode implements the encoder interface.
func (r *AuthRefreshToken) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

// Decode implements the decoder interface.
func (r *AuthRefreshToken) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// =================================================================

// Token represents the user token when requested.
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Encode implements the encoder interface.
func (t Token) Encode() ([]byte, string, error) {
	data, err := json.Marshal(t)
	return data, "application/json", err
}
