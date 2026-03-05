package handler

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/app/handler/openapi"
	"github.com/Housiadas/cerberus/internal/usecase/role_usecase"
	"github.com/Housiadas/cerberus/internal/utils/pntr"
)

func (h *Handler) ListRoles(
	ctx context.Context,
	request openapi.ListRolesRequestObject,
) (openapi.ListRolesResponseObject, error) {
	qp := role_usecase.AppQueryParams{
		Page:    pntr.DerefStr(request.Params.Page),
		Rows:    pntr.DerefStr(request.Params.Rows),
		OrderBy: pntr.DerefStr(request.Params.OrderBy),
		ID:      pntr.DerefStr(request.Params.RoleId),
		Name:    pntr.DerefStr(request.Params.Name),
	}

	result, err := h.Usecase.Role.Query(ctx, qp)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}

	return openapi.ListRoles200JSONResponse{
		Data:     new(result.Data),
		Metadata: new(result.Metadata),
	}, nil
}

func (h *Handler) CreateRole(
	ctx context.Context,
	request openapi.CreateRoleRequestObject,
) (openapi.CreateRoleResponseObject, error) {
	role, err := h.Usecase.Role.Create(ctx, *request.Body)
	if err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}

	return openapi.CreateRole200JSONResponse(role), nil
}

func (h *Handler) UpdateRole(
	ctx context.Context,
	request openapi.UpdateRoleRequestObject,
) (openapi.UpdateRoleResponseObject, error) {
	role, err := h.Usecase.Role.Update(ctx, *request.Body, request.RoleId)
	if err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}

	return openapi.UpdateRole200JSONResponse(role), nil
}

func (h *Handler) DeleteRole(
	ctx context.Context,
	request openapi.DeleteRoleRequestObject,
) (openapi.DeleteRoleResponseObject, error) {
	err := h.Usecase.Role.Delete(ctx, request.RoleId)
	if err != nil {
		return nil, fmt.Errorf("delete role: %w", err)
	}

	return openapi.DeleteRole204Response{}, nil
}

func (h *Handler) CreateRolePermission(
	ctx context.Context,
	request openapi.CreateRolePermissionRequestObject,
) (openapi.CreateRolePermissionResponseObject, error) {
	err := h.Usecase.UserRolesPermissions.AddRolePermission(
		ctx,
		request.RoleId,
		request.Body.PermissionId,
	)
	if err != nil {
		return nil, fmt.Errorf("create role permission: %w", err)
	}

	return openapi.CreateRolePermission204Response{}, nil
}

func (h *Handler) DeleteRolePermission(
	ctx context.Context,
	request openapi.DeleteRolePermissionRequestObject,
) (openapi.DeleteRolePermissionResponseObject, error) {
	err := h.Usecase.UserRolesPermissions.RemoveRolePermission(
		ctx,
		request.RoleId,
		request.Params.PermissionId,
	)
	if err != nil {
		return nil, fmt.Errorf("delete role permission: %w", err)
	}

	return openapi.DeleteRolePermission204Response{}, nil
}
