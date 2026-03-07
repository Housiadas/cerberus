package handler

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	ctxPck "github.com/Housiadas/cerberus/internal/utils/context"
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
