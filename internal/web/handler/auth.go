package handler

import (
	"context"
	"errors"
	"fmt"

	ctxPck "github.com/Housiadas/cerberus/internal/context"
	"github.com/Housiadas/cerberus/internal/core/auth"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
)

func (h *Handler) AuthLogin(
	ctx context.Context,
	request openapi.AuthLoginRequestObject,
) (openapi.AuthLoginResponseObject, error) {
	var fieldErrors errs.FieldErrors
	if request.Body.Email == "" {
		fieldErrors.Add("email", errors.New("email is required"))
	}
	if request.Body.Password == "" {
		fieldErrors.Add("password", errors.New("password is required"))
	}
	if len(fieldErrors) > 0 {
		return nil, fieldErrors.ToError()
	}

	token, err := h.svc.auth.Login(ctx, auth.LoginReq{
		Email:    request.Body.Email,
		Password: request.Body.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("auth login: %w", err)
	}

	return openapi.AuthLogin200JSONResponse(toOpenAPIToken(token)), nil
}

func (h *Handler) AuthRegister(
	ctx context.Context,
	request openapi.AuthRegisterRequestObject,
) (openapi.AuthRegisterResponseObject, error) {
	nu, err := parseNewUser(request.Body)
	if err != nil {
		return nil, errs.New(errs.InvalidArgument, errs.CodeValidation, err)
	}

	usr, err := h.svc.user.Create(ctx, nu)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmail) {
			return nil, errs.New(errs.Aborted, errs.CodeUniqueEmail, user.ErrUniqueEmail)
		}

		return nil, fmt.Errorf("auth register: %w", err)
	}

	return openapi.AuthRegister200JSONResponse(toOpenAPIUser(usr)), nil
}

func (h *Handler) AuthLogout(
	ctx context.Context,
	request openapi.AuthLogoutRequestObject,
) (openapi.AuthLogoutResponseObject, error) {
	actorID := ctxPck.GetActorID(ctx)

	err := h.svc.auth.Logout(ctx, actorID, auth.LogoutReq{
		Token: request.Body.RefreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("auth logout: %w", err)
	}

	return openapi.AuthLogout204Response{}, nil
}

func (h *Handler) AuthRefresh(
	ctx context.Context,
	request openapi.AuthRefreshRequestObject,
) (openapi.AuthRefreshResponseObject, error) {
	token, err := h.svc.auth.RefreshAccessToken(ctx, auth.RefreshTokenReq{
		Token: request.Body.RefreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("auth refresh: %w", err)
	}

	return openapi.AuthRefresh200JSONResponse(toOpenAPIToken(token)), nil
}

func (h *Handler) AuthForgotPassword(
	ctx context.Context,
	request openapi.AuthForgotPasswordRequestObject,
) (openapi.AuthForgotPasswordResponseObject, error) {
	err := h.svc.auth.ForgotPassword(ctx, auth.ForgotPasswordReq{
		Email: request.Body.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("auth forgot password: %w", err)
	}

	return openapi.AuthForgotPassword204Response{}, nil
}

func (h *Handler) AuthResetPassword(
	ctx context.Context,
	request openapi.AuthResetPasswordRequestObject,
) (openapi.AuthResetPasswordResponseObject, error) {
	err := h.svc.auth.ResetPassword(ctx, auth.ResetPasswordReq{
		Token:           request.Body.Token,
		OldPassword:     request.Body.OldPassword,
		Password:        request.Body.Password,
		PasswordConfirm: request.Body.PasswordConfirm,
	})
	if err != nil {
		return nil, fmt.Errorf("auth reset password: %w", err)
	}

	return openapi.AuthResetPassword204Response{}, nil
}
