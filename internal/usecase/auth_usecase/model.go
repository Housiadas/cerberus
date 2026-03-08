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

// =================================================================

// ForgotPasswordReq defines the data needed to initiate a password reset.
type ForgotPasswordReq struct {
	Email string `json:"email" validate:"required"`
}

// Validate checks the data in the model is considered clean.
func (f *ForgotPasswordReq) Validate() error {
	err := validation.Check(f)
	if err != nil {
		return fmt.Errorf("forgot password req validation error: %w", err)
	}

	return nil
}

// =================================================================

// ResetPasswordReq defines the data needed to complete a password reset.
type ResetPasswordReq struct {
	Token           string `json:"token"           validate:"required"`
	OldPassword     string `json:"oldPassword"     validate:"required"`
	Password        string `json:"password"        validate:"required"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required"`
}

// Validate checks the data in the model is considered clean.
func (r *ResetPasswordReq) Validate() error {
	err := validation.Check(r)
	if err != nil {
		return fmt.Errorf("reset password req validation error: %w", err)
	}

	return nil
}
