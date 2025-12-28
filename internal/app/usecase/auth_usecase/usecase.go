package auth_usecase

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_roles_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/logger"
)

// Configuration
var (
	// Use strong, random secrets in production (store in env vars)
	accessTokenSecret  = []byte("your-256-bit-access-secret")
	refreshTokenSecret = []byte("your-256-bit-refresh-secret")
	accessTokenExpiry  = 15 * time.Minute
	refreshTokenExpiry = 7 * 24 * time.Hour
)

// Config represents information required to initialize auth.
type Config struct {
	Issuer           string
	Log              *logger.Logger
	UserUsecase      *user_usecase.UseCase
	UserRolesUsecase *user_roles_usecase.UseCase
}

// UseCase is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type UseCase struct {
	issuer           string
	Secrets          Secrets
	parser           *jwt.Parser
	method           jwt.SigningMethod
	log              *logger.Logger
	userUsecase      *user_usecase.UseCase
	userRolesUsecase *user_roles_usecase.UseCase
}

// Claims represent the authorization claims transmitted via a JWT.
type Claims struct {
	jwt.RegisteredClaims
	TokenID string   `json:"jti"` // JWT ID for token revocation
	Roles   []string `json:"roles"`
}

type Secrets struct {
	AccessTokenSecret  []byte
	RefreshTokenSecret []byte
}

// NewUseCase creates a UseCase to support authentication/authorization.
func NewUseCase(cfg Config) *UseCase {
	return &UseCase{
		log:              cfg.Log,
		issuer:           cfg.Issuer,
		userUsecase:      cfg.UserUsecase,
		userRolesUsecase: cfg.UserRolesUsecase,
		Secrets: Secrets{
			AccessTokenSecret:  accessTokenSecret,
			RefreshTokenSecret: refreshTokenSecret,
		},
		method: jwt.GetSigningMethod(jwt.SigningMethodHS256.Name),
		parser: jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name})),
	}
}

// Issuer provides the configured issuer used to authenticate tokens.
func (u *UseCase) Issuer() string {
	return u.issuer
}

func (u *UseCase) CheckExpiredToken(claims Claims) error {
	// Check if the token has expired
	expiredAt := claims.ExpiresAt
	if time.Now().Unix() > expiredAt.Unix() {
		return fmt.Errorf("token has expired")
	}

	return nil
}

// isUserEnabled hits the database and checks the user is not disabled.
func (u *UseCase) isUserEnabled(ctx context.Context, claims Claims) error {
	usr, err := u.userUsecase.QueryByID(ctx, claims.Subject)
	if err != nil {
		return fmt.Errorf("query user: %w", err)
	}

	if !usr.Enabled {
		return ErrUserDisabled
	}

	return nil
}
