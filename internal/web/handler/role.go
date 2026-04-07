package handler

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/role"
	errs "github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/Housiadas/cerberus/internal/sdk/pntr"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

func (h *Handler) ListRoles(
	ctx context.Context,
	request openapi.ListRolesRequestObject,
) (openapi.ListRolesResponseObject, error) {
	cur, err := cursor.Parse(
		pntr.DerefStr(request.Params.Cursor),
		pntr.DerefStr(request.Params.Limit),
	)
	if err != nil {
		return nil, errs.NewFieldErrors("cursor", err)
	}

	filter, err := parseRoleFilter(
		pntr.DerefStr(request.Params.RoleId),
		pntr.DerefStr(request.Params.Name),
	)
	if err != nil {
		return nil, err
	}

	orderBy, err := order.Parse(
		roleOrderByFields(),
		pntr.DerefStr(request.Params.OrderBy),
		roleDefaultOrderBy(),
	)
	if err != nil {
		return nil, errs.NewFieldErrors("order", err)
	}

	roles, err := h.svc.role.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, errs.CodeInternal, "query: %s", err)
	}

	result := cursor.NewResult(roles, cur.Limit(), cur, orderBy,
		func(r role.Role) string { return r.ID().String() },
		func(r role.Role) any { return r.ID().String() },
	)

	return openapi.ListRoles200JSONResponse{
		Data:     new(toOpenAPIRoles(result.Data)),
		Metadata: new(toOpenAPIMetadata(result.Metadata)),
	}, nil
}

func (h *Handler) CreateRole(
	ctx context.Context,
	request openapi.CreateRoleRequestObject,
) (openapi.CreateRoleResponseObject, error) {
	if request.Body.Name == "" {
		return nil, errs.NewFieldErrors("name", role.ErrRequiredName)
	}

	nme, err := name.Parse(request.Body.Name)
	if err != nil {
		return nil, errs.NewFieldErrors("name", err)
	}

	rol, err := h.svc.role.Create(ctx, role.NewRole{Name: nme})
	if err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}

	return openapi.CreateRole200JSONResponse(toOpenAPIRole(rol)), nil
}

//nolint:dupl // update handler pattern shared across domains
func (h *Handler) UpdateRole(
	ctx context.Context,
	request openapi.UpdateRoleRequestObject,
) (openapi.UpdateRoleResponseObject, error) {
	roleUUID, err := uuid.Parse(request.RoleId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	ur, err := parseUpdateRole(request.Body)
	if err != nil {
		return nil, errs.New(errs.InvalidArgument, errs.CodeValidation, err)
	}

	currentRole, err := h.svc.role.QueryByID(ctx, roleUUID)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, errs.CodeInternal, "role query by id: %s", err)
	}

	updRole, err := h.svc.role.Update(ctx, currentRole, ur)
	if err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}

	return openapi.UpdateRole200JSONResponse(toOpenAPIRole(updRole)), nil
}

func (h *Handler) DeleteRole(
	ctx context.Context,
	request openapi.DeleteRoleRequestObject,
) (openapi.DeleteRoleResponseObject, error) {
	roleUUID, err := uuid.Parse(request.RoleId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentRole, err := h.svc.role.QueryByID(ctx, roleUUID)
	if err != nil {
		return nil, errs.Errorf(errs.Internal, errs.CodeInternal, "role query by id: %s", err)
	}

	err = h.svc.role.Delete(ctx, currentRole)
	if err != nil {
		return nil, fmt.Errorf("delete role: %w", err)
	}

	return openapi.DeleteRole204Response{}, nil
}

func (h *Handler) CreateRolePermission(
	ctx context.Context,
	request openapi.CreateRolePermissionRequestObject,
) (openapi.CreateRolePermissionResponseObject, error) {
	roleUUID, err := uuid.Parse(request.RoleId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse role uuid: %s",
			err,
		)
	}

	permUUID, err := uuid.Parse(request.Body.PermissionId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse permission uuid: %s",
			err,
		)
	}

	err = h.svc.rolePermissions.Add(ctx, roleUUID, permUUID)
	if err != nil {
		return nil, fmt.Errorf("create role permission: %w", err)
	}

	return openapi.CreateRolePermission204Response{}, nil
}

func (h *Handler) DeleteRolePermission(
	ctx context.Context,
	request openapi.DeleteRolePermissionRequestObject,
) (openapi.DeleteRolePermissionResponseObject, error) {
	roleUUID, err := uuid.Parse(request.RoleId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse role uuid: %s",
			err,
		)
	}

	permUUID, err := uuid.Parse(request.Params.PermissionId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse permission uuid: %s",
			err,
		)
	}

	err = h.svc.rolePermissions.Remove(ctx, roleUUID, permUUID)
	if err != nil {
		return nil, fmt.Errorf("delete role permission: %w", err)
	}

	return openapi.DeleteRolePermission204Response{}, nil
}

// ---------------------------------------------------------------------------
// Parsing helpers
// ---------------------------------------------------------------------------

func parseUpdateRole(body *openapi.UpdateRole) (role.UpdateRole, error) {
	var ur role.UpdateRole

	if body.Name != nil {
		nme, err := name.Parse(*body.Name)
		if err != nil {
			return role.UpdateRole{}, fmt.Errorf("parse: %w", err)
		}

		ur.Name = &nme
	}

	return ur, nil
}

func parseRoleFilter(id, nameStr string) (role.QueryFilter, error) {
	var (
		fieldErrors errs.FieldErrors
		filter      role.QueryFilter
	)

	if id != "" {
		uid, err := uuid.Parse(id)
		if err != nil {
			fieldErrors.Add("role_id", err)
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
		return role.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}

func roleDefaultOrderBy() order.By {
	return order.NewBy("id", order.ASC)
}

func roleOrderByFields() map[string]string {
	return map[string]string{
		"id": role.OrderByID,
	}
}
