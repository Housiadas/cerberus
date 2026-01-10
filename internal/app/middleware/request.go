package middleware

import (
	"context"
	"net/http"

	ctxPck "github.com/Housiadas/cerberus/internal/common/context"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// RequestID is a middleware that injects uuid as middleware.RequestIDHeader when not present.
func (m *Middleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			u   uuid.UUID
			err error
		)

		ctx := r.Context()
		reqID := r.Header.Get(middleware.RequestIDHeader)

		if reqID == "" {
			u = initializeUUIDV7(ctx, m.Log)
		} else {
			u, err = uuid.Parse(reqID)
			if err != nil {
				m.Log.Info(ctx, "request id parse error", err)
				u = initializeUUIDV7(ctx, m.Log)
			}
		}

		us := u.String()
		ctx = ctxPck.SetRequestID(ctx, us)
		w.Header().Set(middleware.RequestIDHeader, us)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func initializeUUIDV7(ctx context.Context, log logger.Logger) uuid.UUID {
	u, err := uuid.NewV7()
	if err != nil {
		log.Info(ctx, "uuid v7 parse", err)

		return uuid.New()
	}

	return u
}
