package handler

import (
	"context"
	"net/http"

	"github.com/Housiadas/cerberus/internal/app/usecase/permission_usecase"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/Housiadas/cerberus/pkg/web"
)

// Permission godoc
// @Summary      Crete Permission
// @Description  Create a new Permission
// @Tags 		 Permission
// @Accept       json
// @Produce      json
// @Param        request body permission_usecase.NewRole true "Permission data"
// @Success      200  {object}  permission_usecase.Permission
// @Failure      500  {object}  errs.Error
// @Router       /permission [post]
func (h *Handler) permissionCreate(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	var ucRole permission_usecase.NewPermission
	if err := web.Decode(r, &ucRole); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	usr, err := h.UseCase.Permission.Create(ctx, ucRole)
	if err != nil {
		return errs.NewError(err)
	}

	return usr
}

// Permission godoc
// @Summary      Query Roles
// @Description  Search roles
// @Tags		 Roles
// @Accept       json
// @Produce      json
// @Success      200  {object}  permission_usecase.RolePageResult
// @Failure      500  {object}  errs.Error
// @Router       /permission [get]
func (h *Handler) permissionQuery(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	qp := permissionParseQueryParams(r)

	roles, err := h.UseCase.Permission.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return roles
}

// Permission godoc
// @Summary      Update Permission
// @Description  Update an existing Permission
// @Tags 		 Permission
// @Accept       json
// @Produce      json
// @Param        request body permission_usecase.UpdateRole true "Permission data"
// @Success      200  {object}  permission_usecase.Permission
// @Failure      500  {object}  errs.Error
// @Router       /permission/{permission_id} [put]
func (h *Handler) permissionUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Encoder {
	var app permission_usecase.UpdatePermission
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	permission, err := h.UseCase.Permission.Update(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return permission
}

// Permission godoc
// @Summary      Delete a Permission
// @Description  Delete a Permission
// @Tags 		 Permission
// @Accept       json
// @Produce      json
// @Success      204
// @Failure      500  {object}  errs.Error
// @Router       /permission/{permission_id} [delete]
func (h *Handler) permissionDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Encoder {
	if err := h.UseCase.Permission.Delete(ctx); err != nil {
		return errs.NewError(err)
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
