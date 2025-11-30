package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/riandyrn/otelchi"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Routes returns applications router
func (h *Handler) Routes() *chi.Mux {
	mid := h.Web.Middleware
	tran := mid.BeginCommitRollback()

	apiRouter := chi.NewRouter()
	apiRouter.Use(
		mid.Recoverer(),
		mid.RequestID,
		mid.Logger(),
		mid.Otel(),
		middleware.SetHeader("Content-Type", "application/json"),
		middleware.GetHead,
	)

	// v1 routes
	apiRouter.Route("/v1", func(v1 chi.Router) {
		v1.Use(
			mid.ApiVersion("v1"),
			otelchi.Middleware(h.ServiceName, otelchi.WithChiRoutes(v1)),
		)
		v1.Use(cors.Handler(cors.Options{
			AllowedOrigins: h.Cors.AllowedOrigins,
			AllowedMethods: h.Cors.AllowedMethods,
			AllowedHeaders: h.Cors.AllowedHeaders,
			ExposedHeaders: h.Cors.ExposedHeaders,
			MaxAge:         h.Cors.MaxAge,
		}))

		// Roles
		v1.Route("/roles", func(p chi.Router) {
			p.Get("/", h.Web.Res.Respond(h.roleQuery))
			p.Post("/", h.Web.Res.Respond(h.roleCreate))
			p.Put("/{role_id}", h.Web.Res.Respond(h.roleUpdate))
			p.Delete("/{role_id}", h.Web.Res.Respond(h.roleDelete))
		})

		// Permissions
		v1.Route("/permissions", func(p chi.Router) {
			// ...
		})

		// Users
		v1.Route("/users", func(u chi.Router) {
			u.Get("/", h.Web.Res.Respond(h.userQuery))
			u.Post("/", h.Web.Res.Respond(h.userCreate))
			u.Get("/{user_id}", h.Web.Res.Respond(h.userQueryByID))
			u.Put("/{user_id}", h.Web.Res.Respond(h.userUpdate))
			u.Put("/{user_id}/roles/{role_id}", h.Web.Res.Respond(h.updateRole))
			u.Delete("/{user_id}", h.Web.Res.Respond(h.userDelete))
		})

		// Audits
		v1.Route("/audits", func(a chi.Router) {
			a.Get("/", h.Web.Res.Respond(h.auditQuery))
		})

		// Transaction example
		v1.With(tran).Post("/transaction", h.Web.Res.Respond(h.transaction))
	})

	// System Routes
	router := chi.NewRouter()
	router.Get("/readiness", h.Web.Res.Respond(h.readiness))
	router.Get("/liveness", h.Web.Res.Respond(h.liveness))
	router.Get("/swagger/doc.json", h.Swagger)
	router.Handle("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("./doc.json"),
	))

	router.Mount("/api", apiRouter)
	return router
}
