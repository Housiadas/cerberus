package handler

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/permission"
	errs "github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/Housiadas/cerberus/internal/sdk/pntr"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

//nolint:dupl // list handler pattern shared across domains
func (h *Handler) ListPermissions(
	ctx context.Context,
	request openapi.ListPermissionsRequestObject,
) (openapi.ListPermissionsResponseObject, error) {
	cur, err := cursor.Parse(
		pntr.DerefStr(request.Params.Cursor),
		pntr.DerefStr(request.Params.Limit),
	)
	if err != nil {
		return nil, errs.NewFieldErrors("cursor", err)
	}

	filter, err := parsePermissionFilter(
		pntr.DerefStr(request.Params.PermissionId),
		pntr.DerefStr(request.Params.Name),
	)
	if err != nil {
		return nil, err
	}

	orderBy, err := order.Parse(
		permissionOrderByFields(),
		pntr.DerefStr(request.Params.OrderBy),
		permissionDefaultOrderBy(),
	)
	if err != nil {
		return nil, errs.NewFieldErrors("order", err)
	}

	perms, err := h.svc.permission.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, errs.CodeInternal, "query: %s", err)
	}

	result := cursor.NewResult(perms, cur.Limit(), cur, orderBy,
		func(p permission.Permission) string { return p.ID().String() },
		permissionFieldExtractor(orderBy),
	)

	return openapi.ListPermissions200JSONResponse{
		Data:     new(toOpenAPIPermissions(result.Data)),
		Metadata: new(toOpenAPIMetadata(result.Metadata)),
	}, nil
}

func (h *Handler) CreatePermission(
	ctx context.Context,
	request openapi.CreatePermissionRequestObject,
) (openapi.CreatePermissionResponseObject, error) {
	if request.Body.Name == "" {
		return nil, errs.NewFieldErrors("name", permission.ErrRequiredName)
	}

	nme, err := name.Parse(request.Body.Name)
	if err != nil {
		return nil, errs.NewFieldErrors("name", err)
	}

	perm, err := h.svc.permission.Create(ctx, permission.NewPermission{Name: nme})
	if err != nil {
		return nil, fmt.Errorf("create permission: %w", err)
	}

	return openapi.CreatePermission200JSONResponse(toOpenAPIPermission(perm)), nil
}

//nolint:dupl // update handler pattern shared across domains
func (h *Handler) UpdatePermission(
	ctx context.Context,
	request openapi.UpdatePermissionRequestObject,
) (openapi.UpdatePermissionResponseObject, error) {
	permUUID, err := uuid.Parse(request.PermissionId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	up, err := parseUpdatePermission(request.Body)
	if err != nil {
		return nil, errs.New(errs.InvalidArgument, errs.CodeValidation, err)
	}

	currentPerm, err := h.svc.permission.QueryByID(ctx, permUUID)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, errs.CodeInternal, "permission query by id: %s", err)
	}

	updPerm, err := h.svc.permission.Update(ctx, currentPerm, up)
	if err != nil {
		return nil, fmt.Errorf("update permission: %w", err)
	}

	return openapi.UpdatePermission200JSONResponse(toOpenAPIPermission(updPerm)), nil
}

func (h *Handler) DeletePermission(
	ctx context.Context,
	request openapi.DeletePermissionRequestObject,
) (openapi.DeletePermissionResponseObject, error) {
	permUUID, err := uuid.Parse(request.PermissionId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentPerm, err := h.svc.permission.QueryByID(ctx, permUUID)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, errs.CodeInternal, "permission query by id: %s", err)
	}

	err = h.svc.permission.Delete(ctx, currentPerm)
	if err != nil {
		return nil, fmt.Errorf("delete permission: %w", err)
	}

	return openapi.DeletePermission204Response{}, nil
}

// ---------------------------------------------------------------------------
// Parsing helpers
// ---------------------------------------------------------------------------

func parseUpdatePermission(body *openapi.UpdatePermission) (permission.UpdatePermission, error) {
	var up permission.UpdatePermission

	if body.Name != nil {
		nme, err := name.Parse(*body.Name)
		if err != nil {
			return permission.UpdatePermission{}, fmt.Errorf("parse: %w", err)
		}

		up.Name = &nme
	}

	return up, nil
}

func parsePermissionFilter(id, nameStr string) (permission.QueryFilter, error) {
	var (
		fieldErrors errs.FieldErrors
		filter      permission.QueryFilter
	)

	if id != "" {
		uid, err := uuid.Parse(id)
		if err != nil {
			fieldErrors.Add("permission_id", err)
		} else {
			filter.ID = &uid
		}
	}

	if nameStr != "" {
		n, err := name.Parse(nameStr)
		if err != nil {
			fieldErrors.Add("name", err)
		} else {
			filter.Name = &n
		}
	}

	if len(fieldErrors) > 0 {
		return permission.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}

func permissionDefaultOrderBy() order.By {
	return order.NewBy("id", order.ASC)
}

func permissionOrderByFields() map[string]string {
	return map[string]string{
		"id":   permission.OrderByID,
		"name": permission.OrderByName,
	}
}

func permissionFieldExtractor(orderBy order.By) func(permission.Permission) any {
	return func(p permission.Permission) any {
		switch orderBy.Field {
		case permission.OrderByName:
			return p.Name().String()
		default:
			return p.ID().String()
		}
	}
}
