package handler

import (
	"context"
	"net/http"

	"github.com/Housiadas/cerberus/internal/app/usecase/permission_usecase"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

// Permission godoc
// @Summary      Crete Permission
// @Description  Create a new Permission
// @Tags 		 Permissions
// @Accept       json
// @Produce      json
// @Param        request body 	permission_usecase.NewPermission true "Permission data"
// @Success      200  {object}  permission_usecase.Permission
// @Failure      500  {object}  errs.Error
// @Router       /v1/permissions [post]
func (h *Handler) permissionCreate(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	var ucRole permission_usecase.NewPermission
	if err := web.Decode(r, &ucRole); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	usr, err := h.Usecase.Permission.Create(ctx, ucRole)
	if err != nil {
		return errs.AsErr(err)
	}

	return usr
}

// Permission godoc
// @Summary      Query Roles
// @Description  Search roles
// @Tags		 Permissions
// @Accept       json
// @Produce      json
// @Success      200  {object}  permission_usecase.PermissionPageResults
// @Failure      500  {object}  errs.Error
// @Router       /v1/permissions [get]
func (h *Handler) permissionQuery(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	qp := permissionParseQueryParams(r)

	roles, err := h.Usecase.Permission.Query(ctx, qp)
	if err != nil {
		return errs.AsErr(err)
	}

	return roles
}

// Permission godoc
// @Summary      Update Permission
// @Description  Update an existing Permission
// @Tags 		 Permissions
// @Accept       json
// @Produce      json
// @Param        request body 	permission_usecase.UpdatePermission true "Permission data"
// @Success      200  {object}  permission_usecase.Permission
// @Failure      500  {object}  errs.Error
// @Router       /v1/permissions/{permission_id} [put]
func (h *Handler) permissionUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Encoder {
	var app permission_usecase.UpdatePermission
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	permissionID := web.Param(r, "permission_id")
	permission, err := h.Usecase.Permission.Update(ctx, app, permissionID)
	if err != nil {
		return errs.AsErr(err)
	}

	return permission
}

// Permission godoc
// @Summary      Delete a Permission
// @Description  Delete a Permission
// @Tags 		 Permissions
// @Accept       json
// @Produce      json
// @Success      204
// @Failure      500  {object}  errs.Error
// @Router       /v1/permissions/{permission_id} [delete]
func (h *Handler) permissionDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Encoder {
	permissionID := web.Param(r, "permission_id")
	if err := h.Usecase.Permission.Delete(ctx, permissionID); err != nil {
		return errs.AsErr(err)
	}

	return nil
}

func permissionParseQueryParams(r *http.Request) permission_usecase.AppQueryParams {
	values := r.URL.Query()

	return permission_usecase.AppQueryParams{
		ID:      values.Get("permission_id"),
		Name:    values.Get("name"),
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
	}
}
