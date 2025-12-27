package auth_usecase

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_roles_usecase"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/logger"
)

//go:embed secret.pem
var staticSecret []byte // todo: fetch secret from vault

// Config represents information required to initialize auth.
type Config struct {
	Issuer           string
	Log              *logger.Logger
	UserUsecase      *user_usecase.UseCase
	UserRolesUsecase *user_roles_usecase.UseCase
}

// Usecase is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Usecase struct {
	issuer           string
	secret           []byte
	parser           *jwt.Parser
	method           jwt.SigningMethod
	log              *logger.Logger
	userUsecase      *user_usecase.UseCase
	userRolesUsecase *user_roles_usecase.UseCase
}

// NewUseCase creates a Usecase to support authentication/authorization.
func NewUseCase(cfg Config) *Usecase {
	return &Usecase{
		log:              cfg.Log,
		issuer:           cfg.Issuer,
		userUsecase:      cfg.UserUsecase,
		userRolesUsecase: cfg.UserRolesUsecase,
		secret:           staticSecret,
		method:           jwt.GetSigningMethod(jwt.SigningMethodRS256.Name),
		parser:           jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name})),
	}
}

// Issuer provides the configured issuer used to authenticate tokens.
func (u *Usecase) Issuer() string {
	return u.issuer
}

func (u *Usecase) Login(ctx context.Context, authLogin AuthLogin) (Token, error) {
	authUsr := user_usecase.AuthenticateUser{
		Email:    authLogin.Email,
		Password: authLogin.Password,
	}

	usr, err := u.userUsecase.Authenticate(ctx, authUsr)
	if err != nil {
		return Token{}, errs.New(errs.Unauthenticated, err)
	}

	// get user roles name
	roles, err := u.userRolesUsecase.GetUserRolesNames(ctx, usr.ID)
	if err != nil {
		return Token{}, errs.New(errs.NotFound, err)
	}

	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   usr.ID,
			Issuer:    u.Issuer(),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: roles,
	}
	token, err := u.GenerateToken(claims)
	if err != nil {
		return Token{}, fmt.Errorf("generate token: %w", err)
	}

	return Token{
		Token: token,
	}, nil
}

// Verify processes the token to validate the sender's token is valid.
func (u *Usecase) Verify(ctx context.Context, bearerToken string) (Claims, error) {
	if !strings.HasPrefix(bearerToken, "Bearer ") {
		return Claims{}, errors.New("expected authorization header format: Bearer <token>")
	}

	jwtUnverified := bearerToken[7:]

	var claims Claims
	token, err := jwt.ParseWithClaims(jwtUnverified, &claims, func(token *jwt.Token) (interface{}, error) {
		return u.secret, nil
	})
	if err != nil {
		return Claims{}, fmt.Errorf("error parsing token: %w", err)
	}
	if !token.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}

	// Check the database for this user to verify they are still enabled.
	if err := u.isUserEnabled(ctx, claims); err != nil {
		return Claims{}, fmt.Errorf("user not enabled: %w", err)
	}

	return claims, nil
}

// GenerateToken generates a signed JWT token string representing the user Claims.
func (u *Usecase) GenerateToken(claims Claims) (string, error) {
	token := jwt.NewWithClaims(u.method, claims)

	str, err := token.SignedString(u.secret)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return str, nil
}

// isUserEnabled hits the database and checks the user is not disabled.
func (u *Usecase) isUserEnabled(ctx context.Context, claims Claims) error {
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return fmt.Errorf("parsing user ID %q from claims: %w", claims.Subject, err)
	}

	usr, err := u.userUsecase.QueryByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("query user: %w", err)
	}

	if !usr.Enabled {
		return ErrUserDisabled
	}

	return nil
}
