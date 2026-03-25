package handler

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/pntr"
	"github.com/Housiadas/cerberus/internal/usecase/permission_usecase"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
)

func (h *Handler) ListPermissions(
	ctx context.Context,
	request openapi.ListPermissionsRequestObject,
) (openapi.ListPermissionsResponseObject, error) {
	qp := permission_usecase.AppQueryParams{
		Cursor:  pntr.DerefStr(request.Params.Cursor),
		Limit:   pntr.DerefStr(request.Params.Limit),
		OrderBy: pntr.DerefStr(request.Params.OrderBy),
		ID:      pntr.DerefStr(request.Params.PermissionId),
		Name:    pntr.DerefStr(request.Params.Name),
	}

	result, err := h.usecase.permission.Query(ctx, qp)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}

	return openapi.ListPermissions200JSONResponse{
		Data:     new(result.Data),
		Metadata: new(result.Metadata),
	}, nil
}

func (h *Handler) CreatePermission(
	ctx context.Context,
	request openapi.CreatePermissionRequestObject,
) (openapi.CreatePermissionResponseObject, error) {
	perm, err := h.usecase.permission.Create(ctx, *request.Body)
	if err != nil {
		return nil, fmt.Errorf("create permission: %w", err)
	}

	return openapi.CreatePermission200JSONResponse(perm), nil
}

func (h *Handler) UpdatePermission(
	ctx context.Context,
	request openapi.UpdatePermissionRequestObject,
) (openapi.UpdatePermissionResponseObject, error) {
	perm, err := h.usecase.permission.Update(ctx, *request.Body, request.PermissionId)
	if err != nil {
		return nil, fmt.Errorf("update permission: %w", err)
	}

	return openapi.UpdatePermission200JSONResponse(perm), nil
}

func (h *Handler) DeletePermission(
	ctx context.Context,
	request openapi.DeletePermissionRequestObject,
) (openapi.DeletePermissionResponseObject, error) {
	err := h.usecase.permission.Delete(ctx, request.PermissionId)
	if err != nil {
		return nil, fmt.Errorf("delete permission: %w", err)
	}

	return openapi.DeletePermission204Response{}, nil
}
