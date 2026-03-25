package middleware

import (
	"net/http"

	ctxPck "github.com/Housiadas/cerberus/internal/context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// RequestID is a middleware that injects uuid as middleware.RequestIDHeader when not present.
func (m *Middleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqID := r.Header.Get(middleware.RequestIDHeader)

		u, err := uuid.Parse(reqID)
		if err != nil {
			m.log.Info(ctx, "request id parse error", err)

			u = uuid.New()
		}

		uuidStr := u.String()
		ctx = ctxPck.SetRequestID(ctx, uuidStr)
		w.Header().Set(middleware.RequestIDHeader, uuidStr)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
