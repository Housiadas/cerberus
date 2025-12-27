package handler

import (
	"context"
	"net/http"

	"github.com/Housiadas/cerberus/internal/app/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/pkg/errs"
	"github.com/Housiadas/cerberus/pkg/web"
)

// Auth godoc
// @Summary Auth login
// @Description Verify user's credentials
// @Tags	Auth
// @Accept json
// @Produce json
// @Param request body auth_usecase.AuthLogin true "Login data"
// @Success 200 {object} auth_usecase.Token
// @Failure 500 {object} errs.Error
// @Router	/v1/auth/login [post]
func (h *Handler) authLogin(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	var app auth_usecase.AuthLogin
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	token, err := h.UseCase.Auth.Login(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return token
}
