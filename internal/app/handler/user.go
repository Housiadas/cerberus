package handler

import (
	"context"

	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	"github.com/Housiadas/cerberus/internal/usecase/user_usecase"
	"github.com/Housiadas/cerberus/internal/utils/pntr"
)

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

	result, err := h.Usecase.User.Query(ctx, qp)
	if err != nil {
		return nil, err
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
	usr, err := h.Usecase.User.Create(ctx, *request.Body)
	if err != nil {
		return nil, err
	}

	return openapi.CreateUser200JSONResponse(usr), nil
}

func (h *Handler) GetUser(
	ctx context.Context,
	request openapi.GetUserRequestObject,
) (openapi.GetUserResponseObject, error) {
	usr, err := h.Usecase.User.QueryByID(ctx, request.UserId)
	if err != nil {
		return nil, err
	}

	return openapi.GetUser200JSONResponse(usr), nil
}

func (h *Handler) UpdateUser(
	ctx context.Context,
	request openapi.UpdateUserRequestObject,
) (openapi.UpdateUserResponseObject, error) {
	updUser, err := h.Usecase.User.Update(ctx, *request.Body, request.UserId)
	if err != nil {
		return nil, err
	}

	return openapi.UpdateUser200JSONResponse(updUser), nil
}

func (h *Handler) DeleteUser(
	ctx context.Context,
	request openapi.DeleteUserRequestObject,
) (openapi.DeleteUserResponseObject, error) {
	err := h.Usecase.User.Delete(ctx, request.UserId)
	if err != nil {
		return nil, err
	}

	return openapi.DeleteUser204Response{}, nil
}

func (h *Handler) CreateUserRole(
	ctx context.Context,
	request openapi.CreateUserRoleRequestObject,
) (openapi.CreateUserRoleResponseObject, error) {
	usr, err := h.Usecase.User.Create(ctx, *request.Body)
	if err != nil {
		return nil, err
	}

	return openapi.CreateUserRole200JSONResponse(usr), nil
}

func (h *Handler) DeleteUserRole(
	ctx context.Context,
	request openapi.DeleteUserRoleRequestObject,
) (openapi.DeleteUserRoleResponseObject, error) {
	err := h.Usecase.User.Delete(ctx, request.UserId)
	if err != nil {
		return nil, err
	}

	return openapi.DeleteUserRole204Response{}, nil
}
