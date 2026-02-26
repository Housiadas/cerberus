package auth_usecase

import (
	"fmt"

	"github.com/Housiadas/cerberus/internal/utils/validation"
)

// LoginReq defines the data needed to authenticate a user.
type LoginReq struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required"`
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
