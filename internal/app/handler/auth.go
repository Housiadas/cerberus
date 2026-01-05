package handler

import (
	"context"
	"net/http"

	"github.com/Housiadas/cerberus/internal/app/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	ctxPck "github.com/Housiadas/cerberus/internal/common/context"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
)

// Auth godoc
// @Summary Auth login
// @Description Validate user's credentials
// @Tags	Auth
// @Accept 	json
// @Produce json
// @Param request body auth_usecase.LoginReq true "Login data"
// @Success 200 {object} auth_usecase.Token
// @Failure 500 {object} errs.Error
// @Router	/v1/auth/login [post]
func (h *Handler) authLogin(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	var req auth_usecase.LoginReq
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	token, err := h.Usecase.Auth.Login(ctx, req)
	if err != nil {
		return errs.AsErr(err)
	}

	return token
}

// Auth godoc
// @Summary Auth register
// @Description Register a new user
// @Tags	Auth
// @Accept json
// @Produce json
// @Param request body user_usecase.NewUser true "Login data"
// @Success 200 {object} user_usecase.User
// @Failure 500 {object} errs.Error
// @Router	/v1/auth/register [post]
func (h *Handler) authRegister(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	var req user_usecase.NewUser
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	usr, err := h.Usecase.User.Create(ctx, req)
	if err != nil {
		return errs.AsErr(err)
	}

	return usr
}

// Auth godoc
// @Summary      Logout a user
// @Description  Logout a user
// @Tags 		 Auth
// @Accept       json
// @Produce      json
// @Success      204
// @Failure      500  {object}  errs.Error
// @Router       /v1/auth/logout [post]
func (h *Handler) authLogout(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	var req auth_usecase.LogoutReq
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	claims := ctxPck.GetClaims(ctx)
	err := h.Usecase.Auth.Logout(ctx, claims.Subject, req)
	if err != nil {
		return errs.AsErr(err)
	}

	return nil
}

// Auth godoc
// @Summary 		Auth refresh
// @Description		RefreshAccessToken user's JWT tokens
// @Tags			Auth
// @Accept			json
// @Produce			json
// @Param request body	auth_usecase.RefreshTokenReq true "RefreshAccessToken data"
// @Success	200 {object} auth_usecase.Token
// @Failure	500 {object} errs.Error
// @Router	/v1/auth/refresh [post]
func (h *Handler) authRefresh(ctx context.Context, _ http.ResponseWriter, r *http.Request) web.Encoder {
	var req auth_usecase.RefreshTokenReq
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	token, err := h.Usecase.Auth.RefreshAccessToken(ctx, req)
	if err != nil {
		return errs.AsErr(err)
	}

	return token
}
