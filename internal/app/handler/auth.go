package handler

import (
	"context"

	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	ctxPck "github.com/Housiadas/cerberus/internal/utils/context"
)

func (h *Handler) AuthLogin(
	ctx context.Context,
	request openapi.AuthLoginRequestObject,
) (openapi.AuthLoginResponseObject, error) {
	token, err := h.Usecase.Auth.Login(ctx, *request.Body)
	if err != nil {
		return nil, err
	}

	return openapi.AuthLogin200JSONResponse(token), nil
}

func (h *Handler) AuthRegister(
	ctx context.Context,
	request openapi.AuthRegisterRequestObject,
) (openapi.AuthRegisterResponseObject, error) {
	usr, err := h.Usecase.User.Create(ctx, *request.Body)
	if err != nil {
		return nil, err
	}

	return openapi.AuthRegister200JSONResponse(usr), nil
}

func (h *Handler) AuthLogout(
	ctx context.Context,
	request openapi.AuthLogoutRequestObject,
) (openapi.AuthLogoutResponseObject, error) {
	claims := ctxPck.GetClaims(ctx)

	err := h.Usecase.Auth.Logout(ctx, claims.Subject, *request.Body)
	if err != nil {
		return nil, err
	}

	return openapi.AuthLogout204Response{}, nil
}

func (h *Handler) AuthRefresh(
	ctx context.Context,
	request openapi.AuthRefreshRequestObject,
) (openapi.AuthRefreshResponseObject, error) {
	token, err := h.Usecase.Auth.RefreshAccessToken(ctx, *request.Body)
	if err != nil {
		return nil, err
	}

	return openapi.AuthRefresh200JSONResponse(token), nil
}
