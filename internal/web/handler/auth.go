package handler

import (
	"context"
	"fmt"

	ctxPck "github.com/Housiadas/cerberus/internal/context"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
)

func (h *Handler) AuthLogin(
	ctx context.Context,
	request openapi.AuthLoginRequestObject,
) (openapi.AuthLoginResponseObject, error) {
	token, err := h.usecase.auth.Login(ctx, *request.Body)
	if err != nil {
		return nil, fmt.Errorf("auth login: %w", err)
	}

	return openapi.AuthLogin200JSONResponse(token), nil
}

func (h *Handler) AuthRegister(
	ctx context.Context,
	request openapi.AuthRegisterRequestObject,
) (openapi.AuthRegisterResponseObject, error) {
	usr, err := h.usecase.user.Create(ctx, *request.Body)
	if err != nil {
		return nil, fmt.Errorf("auth register: %w", err)
	}

	return openapi.AuthRegister200JSONResponse(usr), nil
}

func (h *Handler) AuthLogout(
	ctx context.Context,
	request openapi.AuthLogoutRequestObject,
) (openapi.AuthLogoutResponseObject, error) {
	actorID := ctxPck.GetActorID(ctx)

	err := h.usecase.auth.Logout(ctx, actorID, *request.Body)
	if err != nil {
		return nil, fmt.Errorf("auth logout: %w", err)
	}

	return openapi.AuthLogout204Response{}, nil
}

func (h *Handler) AuthRefresh(
	ctx context.Context,
	request openapi.AuthRefreshRequestObject,
) (openapi.AuthRefreshResponseObject, error) {
	token, err := h.usecase.auth.RefreshAccessToken(ctx, *request.Body)
	if err != nil {
		return nil, fmt.Errorf("auth refresh: %w", err)
	}

	return openapi.AuthRefresh200JSONResponse(token), nil
}

func (h *Handler) AuthForgotPassword(
	ctx context.Context,
	request openapi.AuthForgotPasswordRequestObject,
) (openapi.AuthForgotPasswordResponseObject, error) {
	err := h.usecase.auth.ForgotPassword(ctx, *request.Body)
	if err != nil {
		return nil, fmt.Errorf("auth forgot password: %w", err)
	}

	return openapi.AuthForgotPassword204Response{}, nil
}

func (h *Handler) AuthResetPassword(
	ctx context.Context,
	request openapi.AuthResetPasswordRequestObject,
) (openapi.AuthResetPasswordResponseObject, error) {
	err := h.usecase.auth.ResetPassword(ctx, *request.Body)
	if err != nil {
		return nil, fmt.Errorf("auth reset password: %w", err)
	}

	return openapi.AuthResetPassword204Response{}, nil
}
