package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/core/email_notification_outbox"
	"github.com/Housiadas/cerberus/internal/core/refresh_token"
	"github.com/Housiadas/cerberus/internal/core/reset_token"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/core/user_roles_permissions"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
	Log                        *logger.Service
	UserService                *user.Service
	RefreshTokenService        *refresh_token.Service
	ResetTokenService          *reset_token.Service
	EmailNotificationOutboxSvc *email_notification_outbox.Service
	UserRolesPermissions       *user_roles_permissions.Service
}

// Service is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Service struct {
	issuer                     string
	secret                     []byte
	frontendURL                string
	parser                     *jwt.Parser
	log                        *logger.Service
	tx                         pgsql.Beginner
	method                     jwt.SigningMethod
	userService                *user.Service
	refreshTokenService        *refresh_token.Service
	resetTokenSvc              *reset_token.Service
	emailNotificationOutboxSvc *email_notification_outbox.Service
	userRolesPermissions       *user_roles_permissions.Service
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

// NewService creates a Service to support authentication/authorization.
func NewService(cfg Config) *Service {
	return &Service{
		tx:          cfg.DB,
		log:         cfg.Log,
		issuer:      cfg.Issuer,
		frontendURL: cfg.FrontendURL,
		secret:      cfg.AccessTokenSecret,
		method:      jwt.GetSigningMethod(jwt.SigningMethodHS256.Name),
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		),
		userService:                cfg.UserService,
		refreshTokenService:        cfg.RefreshTokenService,
		resetTokenSvc:              cfg.ResetTokenService,
		emailNotificationOutboxSvc: cfg.EmailNotificationOutboxSvc,
		userRolesPermissions:       cfg.UserRolesPermissions,
	}
}

// Issuer provides the configured issuer used to authenticate tokens.
func (s *Service) Issuer() string {
	return s.issuer
}

// CheckExpiredToken validates the expiration claim of a JWT.
func (s *Service) CheckExpiredToken(claims Claims) error {
	expiredAt := claims.ExpiresAt
	if time.Now().Unix() > expiredAt.Unix() {
		return ErrExpiredToken
	}

	return nil
}

// isUserEnabled hits the database and checks the user is not disabled.
func (s *Service) isUserEnabled(ctx context.Context, claims Claims) error {
	userUUID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return fmt.Errorf("parse user uuid: %w", err)
	}

	usr, err := s.userService.QueryByID(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("could not query user: %w", err)
	}

	if !usr.Enabled() {
		return ErrUserDisabled
	}

	return nil
}
