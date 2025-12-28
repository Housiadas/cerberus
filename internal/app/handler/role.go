package handler

import (
	"context"
	"net/http"

	"github.com/Housiadas/cerberus/internal/app/usecase/role_usecase"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/Housiadas/cerberus/pkg/web"
)

// Role godoc
// @Summary      Crete Role
// @Description  Create a new role
// @Tags 		 Role
// @Accept       json
// @Produce      json
// @Param        request body role_usecase.NewRole true "Role data"
// @Success      200  {object}  role_usecase.Role
// @Failure      500  {object}  errs.Error
// @Router       /role [post]
func (h *Handler) roleCreate(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	var ucRole role_usecase.NewRole
	if err := web.Decode(r, &ucRole); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	usr, err := h.UseCase.Role.Create(ctx, ucRole)
	if err != nil {
		return errs.NewError(err)
	}

	return usr
}

func (h *Handler) rolePermissionCreate(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	var ucRole role_usecase.NewRole
	if err := web.Decode(r, &ucRole); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	usr, err := h.UseCase.Role.Create(ctx, ucRole)
	if err != nil {
		return errs.NewError(err)
	}

	return usr
}

// Role godoc
// @Summary      Query Roles
// @Description  Search roles
// @Tags		 Roles
// @Accept       json
// @Produce      json
// @Success      200  {object}  role_usecase.RolePageResult
// @Failure      500  {object}  errs.Error
// @Router       /role [get]
func (h *Handler) roleQuery(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	qp := roleParseQueryParams(r)

	roles, err := h.UseCase.Role.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return roles
}

// Role godoc
// @Summary      Update Role
// @Description  Update an existing role
// @Tags 		 Role
// @Accept       json
// @Produce      json
// @Param        request body role_usecase.UpdateRole true "Role data"
// @Success      200  {object}  role_usecase.Role
// @Failure      500  {object}  errs.Error
// @Router       /role/{role_id} [put]
func (h *Handler) roleUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Encoder {
	var res role_usecase.UpdateRole
	if err := web.Decode(r, &res); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	roleID := web.Param(r, "role_id")
	role, err := h.UseCase.Role.Update(ctx, res, roleID)
	if err != nil {
		return errs.NewError(err)
	}

	return role
}

// Role godoc
// @Summary      Delete a role
// @Description  Delete a role
// @Tags 		 Role
// @Accept       json
// @Produce      json
// @Success      204
// @Failure      500  {object}  errs.Error
// @Router       /role/{role_id} [delete]
func (h *Handler) roleDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) web.Encoder {
	roleID := web.Param(r, "role_id")
	if err := h.UseCase.Role.Delete(ctx, roleID); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func roleParseQueryParams(r *http.Request) role_usecase.AppQueryParams {
	values := r.URL.Query()

	return role_usecase.AppQueryParams{
		ID:      values.Get("role_id"),
		Name:    values.Get("name"),
		Page:    values.Get("page"),
		Rows:    values.Get("rows"),
		OrderBy: values.Get("orderBy"),
	}
}
