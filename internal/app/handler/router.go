package handler

import (
	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/riandyrn/otelchi"
)

// Routes returns applications router.
func (h *Handler) Routes() *chi.Mux {
	m := h.Middleware

	router := chi.NewRouter()
	router.Use(
		m.Recoverer(),
		m.RequestID,
		m.Logger(),
		m.Otel(),
		middleware.SetHeader(ContentTypeKey, ContentTypeJSON),
		middleware.GetHead,
		cors.Handler(cors.Options{
			AllowedOrigins: h.Cors.AllowedOrigins,
			AllowedMethods: h.Cors.AllowedMethods,
			AllowedHeaders: h.Cors.AllowedHeaders,
			ExposedHeaders: h.Cors.ExposedHeaders,
			MaxAge:         h.Cors.MaxAge,
		}),
		otelchi.Middleware(h.ServiceName, otelchi.WithChiRoutes(router)),
	)

	// order matter, first goes auth, then permissions etc.
	si := openapi.NewStrictHandlerWithOptions(h, []openapi.StrictMiddlewareFunc{
		m.Validation(),
		m.Permission(),
		m.Authenticate(),
	}, openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  requestErrorHandler,
		ResponseErrorHandlerFunc: responseErrorHandler,
	})

	openapi.HandlerFromMux(si, router)

	return router
}
