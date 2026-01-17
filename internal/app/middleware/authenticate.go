package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	ctxPck "github.com/Housiadas/cerberus/internal/common/context"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

// AuthenticateBearer is a middleware function that checks authentication,
// validates user credentials, and attach user data to the request context.
func (m *Middleware) AuthenticateBearer() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			bearerToken := r.Header.Get("Authorization")
			if !strings.HasPrefix(bearerToken, "Bearer ") {
				err := errs.New(errs.Unauthenticated, ErrInvalidAuthHeader)
				m.Error(w, err, http.StatusUnauthorized)

				return
			}

			jwtUnverified := bearerToken[7:]

			resp, err := m.UseCase.Auth.Validate(ctx, jwtUnverified)
			if err != nil {
				m.Error(w, err, http.StatusUnauthorized)

				return
			}

			ctx = ctxPck.SetClaims(ctx, resp)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthenticateBasic processes basic authentication logic.
func (m *Middleware) AuthenticateBasic() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authorizationHeader := r.Header.Get("Authorization")

			email, pass, ok := parseBasicAuth(authorizationHeader)
			if !ok {
				err := errs.New(errs.Unauthenticated, ErrInvalidBasicAuth)
				m.Error(w, err, http.StatusUnauthorized)

				return
			}

			authUsr := user_usecase.AuthenticateUser{
				Email:    email,
				Password: pass,
			}

			_, err := m.UseCase.User.Authenticate(ctx, authUsr)
			if err != nil {
				m.Error(w, err, http.StatusUnauthorized)

				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func parseBasicAuth(auth string) (string, string, bool) {
	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Basic" {
		return "", "", false
	}

	c, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", false
	}

	username, password, ok := strings.Cut(string(c), ":")
	if !ok {
		return "", "", false
	}

	return username, password, true
}
