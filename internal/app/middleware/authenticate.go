package middleware

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	"github.com/Housiadas/cerberus/internal/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_usecase"
	ctxPck "github.com/Housiadas/cerberus/internal/utils/context"
	"github.com/Housiadas/cerberus/internal/utils/errs"
)

// publicOperations lists operations that do not require authentication.
var publicOperations = map[string]struct{}{
	"AuthLogin":    {},
	"AuthRegister": {},
	"AuthRefresh":  {},
	"Readiness":    {},
	"Liveness":     {},
}

// AuthStrict applies bearer token authentication for protected operations.
func AuthStrict(authUsecase *auth_usecase.UseCase) openapi.StrictMiddlewareFunc {
	return func(
		f openapi.StrictHandlerFunc,
		operationID string,
	) openapi.StrictHandlerFunc {
		if _, ok := publicOperations[operationID]; ok {
			return f
		}

		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			bearerToken := r.Header.Get("Authorization")
			if !strings.HasPrefix(bearerToken, "Bearer ") {
				return nil, errs.New(errs.Unauthenticated, ErrInvalidAuthHeader)
			}

			jwtUnverified := bearerToken[7:]

			resp, err := authUsecase.Validate(ctx, jwtUnverified)
			if err != nil {
				return nil, errs.New(errs.Unauthenticated, err)
			}

			ctx = ctxPck.SetClaims(ctx, resp)

			return f(ctx, w, r.WithContext(ctx), request)
		}
	}
}

// AuthenticateBasic processes basic authentication logic.
// needs to be changed to openapi-codegen format.
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
