package middleware

import (
	"context"
	"net/http"

	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	ctxPck "github.com/Housiadas/cerberus/internal/utils/context"
	"github.com/Housiadas/cerberus/internal/utils/errs"
)

// operationPermissions maps each operationID to its required permission.
// Operations not listed here do not require permission checks.
var operationPermissions = map[string]string{
	"GetUser":              "user:read",
	"ListUsers":            "user:read:all",
	"CreateUser":           "user:write",
	"UpdateUser":           "user:write",
	"DeleteUser":           "user:delete",
	"ListRoles":            "role:read:all",
	"CreateRole":           "role:write",
	"UpdateRole":           "role:write",
	"DeleteRole":           "role:delete",
	"CreateRolePermission": "role:write",
	"CreateUserRole":       "user:write",
	"DeleteUserRole":       "user:role:delete",
	"ListPermissions":      "permission:read",
	"CreatePermission":     "permission:write",
	"UpdatePermission":     "permission:write",
	"DeletePermission":     "permission:delete",
	"ListAudits":           "audit:read",
}

// Permission checks if the authenticated user has the required permission for the operation.
func (m *Middleware) Permission() openapi.StrictMiddlewareFunc {
	return func(
		f openapi.StrictHandlerFunc,
		operationID string,
	) openapi.StrictHandlerFunc {
		permissionName, ok := operationPermissions[operationID]
		if !ok {
			return f
		}

		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			claims := ctxPck.GetClaims(ctx)
			userID := claims.Subject

			hasPermission, err := m.UseCase.UserRolesPermissions.HasPermission(
				ctx,
				userID,
				permissionName,
			)
			if err != nil {
				m.Log.Error(ctx, "error checking permissions", err)

				return nil, errs.New(errs.Internal, ErrCheckingPermission)
			}

			if !hasPermission {
				m.Log.Info(ctx, "access denied",
					"user_id", userID,
					"permission", permissionName,
					"operation", operationID,
				)

				return nil, errs.New(errs.PermissionDenied, ErrPermissionDenied)
			}

			return f(ctx, w, r, request)
		}
	}
}
