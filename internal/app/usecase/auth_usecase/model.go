package auth_usecase

import (
	"encoding/json"

	"github.com/Housiadas/cerberus/internal/common/validation"
)

// LoginReq defines the data needed to authenticate a user.
type LoginReq struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Encode implements the encoder interface.
func (l *LoginReq) Encode() ([]byte, string, error) {
	data, err := json.Marshal(l)
	return data, "application/json", err
}

// Decode implements the decoder interface.
func (l *LoginReq) Decode(data []byte) error {
	return json.Unmarshal(data, l)
}

// Validate checks the data in the model is considered clean.
func (l *LoginReq) Validate() error {
	if err := validation.Check(l); err != nil {
		return err
	}
	return nil
}

// =================================================================

type RefreshTokenReq struct {
	Token string `json:"refresh_token" validate:"required"`
}

// Encode implements the encoder interface.
func (r *RefreshTokenReq) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

// Decode implements the decoder interface.
func (r *RefreshTokenReq) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// Validate checks the data in the model is considered clean.
func (r *RefreshTokenReq) Validate() error {
	if err := validation.Check(r); err != nil {
		return err
	}
	return nil
}

// =================================================================

type LogoutReq struct {
	Token string `json:"refresh_token" validate:"required"`
}

// Encode implements the encoder interface.
func (r *LogoutReq) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}

// Decode implements the decoder interface.
func (r *LogoutReq) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

// Validate checks the data in the model is considered clean.
func (l *LogoutReq) Validate() error {
	if err := validation.Check(l); err != nil {
		return err
	}
	return nil
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
