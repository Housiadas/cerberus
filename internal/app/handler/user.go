package handler

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	"github.com/Housiadas/cerberus/internal/usecase/user_usecase"
	ctxPck "github.com/Housiadas/cerberus/internal/utils/context"
	"github.com/Housiadas/cerberus/internal/utils/pntr"
)

func (h *Handler) GetMe(
	ctx context.Context,
	_ openapi.GetMeRequestObject,
) (openapi.GetMeResponseObject, error) {
	actorID := ctxPck.GetActorID(ctx)

	usr, err := h.usecase.user.QueryByID(ctx, actorID)
	if err != nil {
		return nil, fmt.Errorf("get me: %w", err)
	}

	return openapi.GetMe200JSONResponse(usr), nil
}

func (h *Handler) UpdateMe(
	ctx context.Context,
	request openapi.UpdateMeRequestObject,
) (openapi.UpdateMeResponseObject, error) {
	actorID := ctxPck.GetActorID(ctx)

	updUser, err := h.usecase.user.UpdateMe(ctx, *request.Body, actorID)
	if err != nil {
		return nil, fmt.Errorf("update me: %w", err)
	}

	return openapi.UpdateMe200JSONResponse(updUser), nil
}

func (h *Handler) ListUsers(
	ctx context.Context,
	request openapi.ListUsersRequestObject,
) (openapi.ListUsersResponseObject, error) {
	qp := user_usecase.AppQueryParams{
		Page:             pntr.DerefStr(request.Params.Page),
		Rows:             pntr.DerefStr(request.Params.Rows),
		OrderBy:          pntr.DerefStr(request.Params.OrderBy),
		ID:               pntr.DerefStr(request.Params.UserId),
		Name:             pntr.DerefStr(request.Params.Name),
		Email:            pntr.DerefStr(request.Params.Email),
		StartCreatedDate: pntr.DerefStr(request.Params.StartCreatedDate),
		EndCreatedDate:   pntr.DerefStr(request.Params.EndCreatedDate),
	}

	result, err := h.usecase.user.Query(ctx, qp)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return openapi.ListUsers200JSONResponse{
		Data:     new(result.Data),
		Metadata: new(result.Metadata),
	}, nil
}

func (h *Handler) CreateUser(
	ctx context.Context,
	request openapi.CreateUserRequestObject,
) (openapi.CreateUserResponseObject, error) {
	usr, err := h.usecase.user.Create(ctx, *request.Body)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return openapi.CreateUser200JSONResponse(usr), nil
}

func (h *Handler) GetUser(
	ctx context.Context,
	request openapi.GetUserRequestObject,
) (openapi.GetUserResponseObject, error) {
	usr, err := h.usecase.user.QueryByID(ctx, request.UserId)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return openapi.GetUser200JSONResponse(usr), nil
}

func (h *Handler) UpdateUser(
	ctx context.Context,
	request openapi.UpdateUserRequestObject,
) (openapi.UpdateUserResponseObject, error) {
	updUser, err := h.usecase.user.Update(ctx, *request.Body, request.UserId)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return openapi.UpdateUser200JSONResponse(updUser), nil
}

func (h *Handler) DeleteUser(
	ctx context.Context,
	request openapi.DeleteUserRequestObject,
) (openapi.DeleteUserResponseObject, error) {
	err := h.usecase.user.Delete(ctx, request.UserId)
	if err != nil {
		return nil, fmt.Errorf("delete user: %w", err)
	}

	return openapi.DeleteUser204Response{}, nil
}

func (h *Handler) CreateUserRole(
	ctx context.Context,
	request openapi.CreateUserRoleRequestObject,
) (openapi.CreateUserRoleResponseObject, error) {
	err := h.usecase.userRoles.AddUserRole(ctx, request.UserId, request.Body.RoleId)
	if err != nil {
		return nil, fmt.Errorf("create user role: %w", err)
	}

	return openapi.CreateUserRole204Response{}, nil
}

func (h *Handler) DeleteUserRole(
	ctx context.Context,
	request openapi.DeleteUserRoleRequestObject,
) (openapi.DeleteUserRoleResponseObject, error) {
	err := h.usecase.userRoles.RemoveUserRole(ctx, request.UserId, request.Params.RoleId)
	if err != nil {
		return nil, fmt.Errorf("delete user role: %w", err)
	}

	return openapi.DeleteUserRole204Response{}, nil
}
