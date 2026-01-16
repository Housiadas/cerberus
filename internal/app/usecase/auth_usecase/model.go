package auth_usecase

import (
	"encoding/json"
	"fmt"

	"github.com/Housiadas/cerberus/internal/common/validation"
)

// LoginReq defines the data needed to authenticate a user.
type LoginReq struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Encode implements the encoder interface.
func (l *LoginReq) Encode() ([]byte, string, error) {
	data, err := json.Marshal(l)
	if err != nil {
		return nil, web.ContentTypeJSON, fmt.Errorf("login req encode err %w", err)
	}

	return data, web.ContentTypeJSON, nil
}

// Decode implements the decoder interface.
func (l *LoginReq) Decode(data []byte) error {
	err := json.Unmarshal(data, l)
	if err != nil {
		return fmt.Errorf("login req decode err %w", err)
	}

	return nil
}

// Validate checks the data in the model is considered clean.
func (l *LoginReq) Validate() error {
	err := validation.Check(l)
	if err != nil {
		return fmt.Errorf("login req encode err %w", err)
	}

	return nil
}

// =================================================================

type RefreshTokenReq struct {
	Token string `json:"refreshToken" validate:"required"`
}

// Encode implements the encoder interface.
func (r *RefreshTokenReq) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, web.ContentTypeJSON, fmt.Errorf("refresh token req encode error: %w", err)
	}

	return data, web.ContentTypeJSON, nil
}

// Decode implements the decoder interface.
func (r *RefreshTokenReq) Decode(data []byte) error {
	err := json.Unmarshal(data, r)

	return fmt.Errorf("refresh token req decode error: %w", err)
}

// Validate checks the data in the model is considered clean.
func (r *RefreshTokenReq) Validate() error {
	err := validation.Check(r)
	if err != nil {
		return fmt.Errorf("refresh token validation error: %w", err)
	}

	return nil
}

// =================================================================

type LogoutReq struct {
	Token string `json:"refreshToken" validate:"required"`
}

// Encode implements the encoder interface.
func (l *LogoutReq) Encode() ([]byte, string, error) {
	data, err := json.Marshal(l)
	if err != nil {
		return nil, web.ContentTypeJSON, fmt.Errorf("logout req encode error: %w", err)
	}

	return data, web.ContentTypeJSON, nil
}

// Decode implements the decoder interface.
func (l *LogoutReq) Decode(data []byte) error {
	err := json.Unmarshal(data, l)

	return fmt.Errorf("logout req decode error: %w", err)
}

// Validate checks the data in the model is considered clean.
func (l *LogoutReq) Validate() error {
	err := validation.Check(l)
	if err != nil {
		return fmt.Errorf("logout req validation error: %w", err)
	}

	return nil
}

// =================================================================

// Token represents the user token when requested.
type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// Encode implements the encoder interface.
func (t Token) Encode() ([]byte, string, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return nil, web.ContentTypeJSON, fmt.Errorf("token encode error: %w", err)
	}

	return data, web.ContentTypeJSON, nil
}
