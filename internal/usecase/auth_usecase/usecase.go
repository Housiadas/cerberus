package auth_usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/service/email_notification_outbox_service"
	"github.com/Housiadas/cerberus/internal/core/service/reset_token_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/internal/usecase/refresh_token_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_roles_permissions_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenTTL  = 20 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

// Config represents information required to initialize auth.
type Config struct {
	Issuer                     string
	FrontendURL                string
	AccessTokenSecret          []byte
	DB                         pgsql.Beginner
	Log                        logger.Logger
	UserUsecase                *user_usecase.UseCase
	UserService                *user_service.Service
	RefreshTokenUsecase        *refresh_token_usecase.UseCase
	ResetTokenService          *reset_token_service.Service
	EmailNotificationOutboxSvc *email_notification_outbox_service.Service
	UserRolesPermissions       *user_roles_permissions_usecase.UseCase
}

// UseCase is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type UseCase struct {
	issuer                     string
	secret                     []byte
	frontendURL                string
	parser                     *jwt.Parser
	log                        logger.Logger
	tx                         pgsql.Beginner
	method                     jwt.SigningMethod
	userUsecase                *user_usecase.UseCase
	userService                *user_service.Service
	refreshTokenUsecase        *refresh_token_usecase.UseCase
	resetTokenSvc              *reset_token_service.Service
	emailNotificationOutboxSvc *email_notification_outbox_service.Service
	userRolesPermissions       *user_roles_permissions_usecase.UseCase
}

// Permission represents a permission embedded in JWT claims.
type Permission struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Claims represent the authorization claims transmitted via a JWT.
type Claims struct {
	jwt.RegisteredClaims

	TokenID     string       `json:"jti"` // JWT ID for token revocation
	Roles       []string     `json:"roles"`
	Permissions []Permission `json:"permissions"`
}

// NewUseCase creates a UseCase to support authentication/authorization.
func NewUseCase(cfg Config) *UseCase {
	return &UseCase{
		tx:          cfg.DB,
		log:         cfg.Log,
		issuer:      cfg.Issuer,
		frontendURL: cfg.FrontendURL,
		secret:      cfg.AccessTokenSecret,
		method:      jwt.GetSigningMethod(jwt.SigningMethodHS256.Name),
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		),
		userUsecase:                cfg.UserUsecase,
		userService:                cfg.UserService,
		refreshTokenUsecase:        cfg.RefreshTokenUsecase,
		resetTokenSvc:              cfg.ResetTokenService,
		emailNotificationOutboxSvc: cfg.EmailNotificationOutboxSvc,
		userRolesPermissions:       cfg.UserRolesPermissions,
	}
}

// Issuer provides the configured issuer used to authenticate tokens.
func (u *UseCase) Issuer() string {
	return u.issuer
}

// CheckExpiredToken validates the expiration claim of a JWT.
func (u *UseCase) CheckExpiredToken(claims Claims) error {
	expiredAt := claims.ExpiresAt
	if time.Now().Unix() > expiredAt.Unix() {
		return ErrExpiredToken
	}

	return nil
}

// isUserEnabled hits the database and checks the user is not disabled.
func (u *UseCase) isUserEnabled(ctx context.Context, claims Claims) error {
	usr, err := u.userUsecase.QueryByID(ctx, claims.Subject)
	if err != nil {
		return fmt.Errorf("could not query user: %w", err)
	}

	if !usr.Enabled {
		return ErrUserDisabled
	}

	return nil
}
