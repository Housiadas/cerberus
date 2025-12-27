package auth_usecase

import (
	"encoding/json"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represent the authorization claims transmitted via a JWT.
type Claims struct {
	jwt.RegisteredClaims
	Roles []string `json:"roles"`
}

// =================================================================

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

// =================================================================

// Token represents the user token when requested.
type Token struct {
	Token string `json:"token"`
}

// Encode implements the encoder interface.
func (t Token) Encode() ([]byte, string, error) {
	data, err := json.Marshal(t)
	return data, "application/json", err
}

func toToken(v string) Token {
	return Token{
		Token: v,
	}
}
