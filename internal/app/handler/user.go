package handler

import (
	"context"
	"net/http"

	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

// User godoc
//
//	@Summary		Crete User
//	@Description	Create a new user
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		user_usecase.NewUser	true	"User data"
//	@Success		200		{object}	user_usecase.User
//	@Failure		500		{object}	errs.Error
//	@Router			/v1/users [post].
func (h *Handler) userCreate(
	ctx context.Context,
	_ http.ResponseWriter,
	r *http.Request,
) web.Encoder {
	var app user_usecase.NewUser
	if err := web.Decode(r, &app); err != nil {
		return errs.ParseValidationErrors(err)
	}

	usr, err := h.Usecase.User.Create(ctx, app)
	if err != nil {
		return errs.AsErr(err)
	}

	return usr
}

// User godoc
//
//	@Summary		Update User
//	@Description	Update an existing user
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		user_usecase.UpdateUser	true	"User data"
//	@Success		200		{object}	user_usecase.User
//	@Failure		500		{object}	errs.Error
//	@Router			/v1/users/{user_id} [put].
func (h *Handler) userUpdate(
	ctx context.Context,
	_ http.ResponseWriter,
	r *http.Request,
) web.Encoder {
	var res user_usecase.UpdateUser
	if err := web.Decode(r, &res); err != nil {
		return errs.ParseValidationErrors(err)
	}

	userID := web.Param(r, "user_id")

	updUser, err := h.Usecase.User.Update(ctx, res, userID)
	if err != nil {
		return errs.AsErr(err)
	}

	return updUser
}

// User godoc
//
//	@Summary		Delete a user
//	@Description	Delete a user
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Failure		500	{object}	errs.Error
//	@Router			/v1/users/{user_id} [delete].
func (h *Handler) userDelete(
	ctx context.Context,
	_ http.ResponseWriter,
	r *http.Request,
) web.Encoder {
	userID := web.Param(r, "user_id")

	err := h.Usecase.User.Delete(ctx, userID)
	if err != nil {
		return errs.AsErr(err)
	}

	return nil
}

// User godoc
//
//	@Summary		Query Users
//	@Description	Search users in database based on criteria
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	user_usecase.UserPageResult
//	@Failure		500	{object}	errs.Error
//	@Router			/v1/users [get].
func (h *Handler) userQuery(
	ctx context.Context,
	_ http.ResponseWriter,
	r *http.Request,
) web.Encoder {
	qp := userParseQueryParams(r)

	usr, err := h.Usecase.User.Query(ctx, qp)
	if err != nil {
		return errs.AsErr(err)
	}

	return usr
}

// User godoc
//
//	@Summary		Find User by id
//	@Description	Search user in database by id
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	user_usecase.User
//	@Failure		500	{object}	errs.Error
//	@Router			/v1/users/{user_id} [get].
func (h *Handler) userQueryByID(
	ctx context.Context,
	_ http.ResponseWriter,
	r *http.Request,
) web.Encoder {
	userID := web.Param(r, "user_id")

	usr, err := h.Usecase.User.QueryByID(ctx, userID)
	if err != nil {
		return errs.AsErr(err)
	}

	return usr
}

func (h *Handler) userRoleCreate(
	ctx context.Context,
	_ http.ResponseWriter,
	r *http.Request,
) web.Encoder {
	var app user_usecase.NewUser
	if err := web.Decode(r, &app); err != nil {
		return errs.ParseValidationErrors(err)
	}

	usr, err := h.Usecase.User.Create(ctx, app)
	if err != nil {
		return errs.AsErr(err)
	}

	return usr
}

func (h *Handler) userRoleDelete(
	ctx context.Context,
	_ http.ResponseWriter,
	r *http.Request,
) web.Encoder {
	userID := web.Param(r, "user_id")

	err := h.Usecase.User.Delete(ctx, userID)
	if err != nil {
		return errs.AsErr(err)
	}

	return nil
}

func userParseQueryParams(r *http.Request) user_usecase.AppQueryParams {
	values := r.URL.Query()

	return user_usecase.AppQueryParams{
		Page:             values.Get("page"),
		Rows:             values.Get("rows"),
		OrderBy:          values.Get("orderBy"),
		ID:               values.Get("user_id"),
		Name:             values.Get("name"),
		Email:            values.Get("email"),
		StartCreatedDate: values.Get("start_created_date"),
		EndCreatedDate:   values.Get("end_created_date"),
	}
}
