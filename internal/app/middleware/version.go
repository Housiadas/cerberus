package middleware

import (
	"net/http"

	"github.com/Housiadas/cerberus/internal/common/context"
)

func (m *Middleware) APIVersion(version string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.SetAPIVersion(r.Context(), version)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
