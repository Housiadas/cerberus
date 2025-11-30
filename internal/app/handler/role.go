package handler

import (
	"context"
	"net/http"

	"github.com/Housiadas/cerberus/internal/app/usecase/role_usecase"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/Housiadas/cerberus/pkg/web"
)

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

func (h *Handler) roleQuery(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	qp := roleParseQueryParams(r)

	usr, err := h.UseCase.Role.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return usr
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
