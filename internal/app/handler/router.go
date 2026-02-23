package handler

import (
	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	mid "github.com/Housiadas/cerberus/internal/app/middleware"
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
		m.Metrics(),
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

	si := openapi.NewStrictHandlerWithOptions(h, []openapi.StrictMiddlewareFunc{
		mid.Validation(),
		mid.AuthStrict(h.Usecase.Auth),
	}, openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  requestErrorHandler,
		ResponseErrorHandlerFunc: responseErrorHandler,
	})

	openapi.HandlerFromMux(si, router)

	return router
}
