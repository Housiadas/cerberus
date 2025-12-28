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
	authenticate := mid.AuthenticateBearer

	apiRouter := chi.NewRouter()
	apiRouter.Use(
		mid.Recoverer(),
		mid.RequestID,
		mid.Logger(),
		mid.Otel(),
		mid.Metrics(),
		middleware.SetHeader("Content-Type", "application/json"),
		middleware.GetHead,
		cors.Handler(cors.Options{
			AllowedOrigins: h.Cors.AllowedOrigins,
			AllowedMethods: h.Cors.AllowedMethods,
			AllowedHeaders: h.Cors.AllowedHeaders,
			ExposedHeaders: h.Cors.ExposedHeaders,
			MaxAge:         h.Cors.MaxAge,
		}),
	)

	// v1 routes
	apiRouter.Route("/v1", func(v1 chi.Router) {
		v1.Use(
			mid.ApiVersion("v1"),
			otelchi.Middleware(h.ServiceName, otelchi.WithChiRoutes(v1)),
		)

		// Auth
		v1.Route("/auth", func(a chi.Router) {
			a.Post("/login", h.Web.Res.Respond(h.authLogin))
			a.Post("/register", h.Web.Res.Respond(h.authRegister))
			a.With(authenticate()).Post("/refresh", h.Web.Res.Respond(h.authRefresh))
		})

		// Users
		v1.With(authenticate()).Route("/users", func(u chi.Router) {
			u.Get("/", h.Web.Res.Respond(h.userQuery))
			u.Post("/", h.Web.Res.Respond(h.userCreate))
			u.Get("/{user_id}", h.Web.Res.Respond(h.userQueryByID))
			u.Put("/{user_id}", h.Web.Res.Respond(h.userUpdate))
			u.Delete("/{user_id}", h.Web.Res.Respond(h.userDelete))
			u.Post("/{user_id}/role", h.Web.Res.Respond(h.userRoleCreate))
			u.Delete("/{user_id}/role", h.Web.Res.Respond(h.userRoleDelete))
		})

		// Roles
		v1.With(authenticate()).Route("/roles", func(r chi.Router) {
			r.Get("/", h.Web.Res.Respond(h.roleQuery))
			r.Post("/", h.Web.Res.Respond(h.roleCreate))
			r.Put("/{role_id}", h.Web.Res.Respond(h.roleUpdate))
			r.Delete("/{role_id}", h.Web.Res.Respond(h.roleDelete))
			r.Post("/{role_id}/permission", h.Web.Res.Respond(h.rolePermissionCreate))
		})

		// Permissions
		v1.With(authenticate()).Route("/permissions", func(p chi.Router) {
			p.Get("/", h.Web.Res.Respond(h.permissionQuery))
			p.Post("/", h.Web.Res.Respond(h.permissionCreate))
			p.Put("/{permission_id}", h.Web.Res.Respond(h.permissionUpdate))
			p.Delete("/{permission_id}", h.Web.Res.Respond(h.permissionDelete))
		})

		// Audits
		v1.With(authenticate()).Route("/audits", func(a chi.Router) {
			a.Get("/", h.Web.Res.Respond(h.auditQuery))
		})
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
