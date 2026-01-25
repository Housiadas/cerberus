package middleware

import (
	"net/http"

	ctxPck "github.com/Housiadas/cerberus/internal/common/context"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

func (m *Middleware) HasPermission(permissionName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			claims := ctxPck.GetClaims(ctx)

			// check if the user has the permission
			userID := claims.Subject

			hasPermission, err := m.UseCase.UserRolesPermissions.HasPermission(
				ctx,
				userID,
				permissionName,
			)
			if err != nil {
				m.Log.Error(ctx, "error checking permissions", err)
				m.Error(w, err, http.StatusInternalServerError)

				return
			}

			if !hasPermission {
				m.Log.Info(ctx, "access denied",
					"user_id", userID,
					"permission", permissionName,
					"has_permissions", hasPermission,
				)
				m.Error(
					w,
					errs.New(errs.PermissionDenied, ErrPermissionDenied),
					http.StatusForbidden,
				)

				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
