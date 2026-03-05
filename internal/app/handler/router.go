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
	m := h.middleware

	router := chi.NewRouter()
	router.Use(
		m.Recoverer(),
		m.RequestID,
		m.Otel(),
		m.Logger(),
		middleware.SetHeader(ContentTypeKey, ContentTypeJSON),
		middleware.GetHead,
		cors.Handler(cors.Options{
			AllowedOrigins: h.cors.AllowedOrigins,
			AllowedMethods: h.cors.AllowedMethods,
			AllowedHeaders: h.cors.AllowedHeaders,
			ExposedHeaders: h.cors.ExposedHeaders,
			MaxAge:         h.cors.MaxAge,
		}),
		otelchi.Middleware(h.serviceName, otelchi.WithChiRoutes(router)),
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
