package middleware

import (
	"net/http"

	"github.com/Housiadas/cerberus/pkg/metrics"
)

// Metrics updates program counters.
func (m *Middleware) Metrics() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := metrics.Set(r.Context())

			// Create a response recorder to capture the response
			rec := &ResponseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rec, r.WithContext(ctx))

			n := metrics.AddRequests(ctx)
			if n%1000 == 0 {
				metrics.AddGoroutines(ctx)
			}

			_, hasError := rec.ResponseWriter.(error)
			if hasError {
				metrics.AddErrors(ctx)
			}
		})
	}
}
