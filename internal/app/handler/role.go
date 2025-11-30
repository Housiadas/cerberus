package handler

import (
	"context"
	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/Housiadas/cerberus/pkg/web"
	"net/http"
)

// Role godoc
// @Summary      Crete Role
// @Description  Create a new role
// @Tags 		 Role
// @Accept       json
// @Produce      json
// @Param        request body user_usecase.NewUser true "User data"
// @Success      200  {object}  user_usecase.User
// @Failure      500  {object}  errs.Error
// @Router       /user [post]
func (h *Handler) roleCreate(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	var app user_usecase.NewUser
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	usr, err := h.App.User.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return usr
}
