package handler

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	ctxPck "github.com/Housiadas/cerberus/internal/context"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/pntr"
	"github.com/Housiadas/cerberus/internal/types/name"
	"github.com/Housiadas/cerberus/internal/types/password"
	"github.com/Housiadas/cerberus/internal/web/handler/openapi"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/google/uuid"
)

func (h *Handler) GetMe(
	ctx context.Context,
	_ openapi.GetMeRequestObject,
) (openapi.GetMeResponseObject, error) {
	actorID := ctxPck.GetActorID(ctx)

	userUUID, err := uuid.Parse(actorID)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s", err,
		)
	}

	usr, err := h.svc.user.QueryByID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, errs.New(
				errs.NotFound,
				errs.CodeUserNotFound, err,
			)
		}

		return nil, errs.New(
			errs.Internal,
			errs.CodeInternal, err,
		)
	}

	return openapi.GetMe200JSONResponse(toOpenAPIUser(usr)), nil
}

func (h *Handler) UpdateMe(
	ctx context.Context,
	request openapi.UpdateMeRequestObject,
) (openapi.UpdateMeResponseObject, error) {
	actorID := ctxPck.GetActorID(ctx)

	userUUID, err := uuid.Parse(actorID)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s", err,
		)
	}

	uu, err := parseUpdateUser(
		request.Body.Name,
		request.Body.Email,
		request.Body.Department,
		request.Body.Password,
		request.Body.PasswordConfirm,
		nil,
	)
	if err != nil {
		return nil, errs.New(
			errs.InvalidArgument,
			errs.CodeValidation, err,
		)
	}

	currentUsr, err := h.svc.user.QueryByID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, errs.New(
				errs.NotFound,
				errs.CodeUserNotFound, err,
			)
		}

		return nil, errs.New(
			errs.Internal,
			errs.CodeInternal, err,
		)
	}

	updUsr, err := h.svc.user.Update(ctx, currentUsr, uu)
	if err != nil {
		return nil, fmt.Errorf("update me: %w", err)
	}

	return openapi.UpdateMe200JSONResponse(toOpenAPIUser(updUsr)), nil
}

func (h *Handler) ListUsers(
	ctx context.Context,
	request openapi.ListUsersRequestObject,
) (openapi.ListUsersResponseObject, error) {
	cur, err := cursor.Parse(
		pntr.DerefStr(request.Params.Cursor),
		pntr.DerefStr(request.Params.Limit),
	)
	if err != nil {
		return nil, errs.NewFieldErrors("cursor", err)
	}

	filter, err := parseUserFilter(
		pntr.DerefStr(request.Params.UserId),
		pntr.DerefStr(request.Params.Name),
		pntr.DerefStr(request.Params.Email),
		pntr.DerefStr(request.Params.StartCreatedDate),
		pntr.DerefStr(request.Params.EndCreatedDate),
	)
	if err != nil {
		return nil, err
	}

	orderBy, err := order.Parse(userOrderByFields(), pntr.DerefStr(request.Params.OrderBy), userDefaultOrderBy())
	if err != nil {
		return nil, errs.NewFieldErrors("order", err)
	}

	usrs, err := h.svc.user.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return nil, errs.New(
			errs.NotFound,
			errs.CodeUserNotFound,
			err,
		)
	}

	result := cursor.NewResult(usrs, cur.Limit(), cur, orderBy, userIDExtractor, userFieldExtractor(orderBy))

	data := make([]openapi.User, len(result.Data))
	for i, u := range result.Data {
		data[i] = toOpenAPIUser(u)
	}

	return openapi.ListUsers200JSONResponse{
		Data:     &data,
		Metadata: new(toOpenAPIMetadata(result.Metadata)),
	}, nil
}

func (h *Handler) CreateUser(
	ctx context.Context,
	request openapi.CreateUserRequestObject,
) (openapi.CreateUserResponseObject, error) {
	nu, err := parseNewUser(request.Body)
	if err != nil {
		return nil, errs.New(errs.InvalidArgument, errs.CodeValidation, err)
	}

	usr, err := h.svc.user.Create(ctx, nu)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmail) {
			return nil, errs.New(
				errs.Aborted,
				errs.CodeUniqueEmail,
				user.ErrUniqueEmail,
			)
		}

		return nil, errs.New(
			errs.Internal,
			errs.CodeInternal, err,
		)
	}

	return openapi.CreateUser200JSONResponse(toOpenAPIUser(usr)), nil
}

func (h *Handler) GetUser(
	ctx context.Context,
	request openapi.GetUserRequestObject,
) (openapi.GetUserResponseObject, error) {
	userUUID, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s", err,
		)
	}

	usr, err := h.svc.user.QueryByID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, errs.New(
				errs.NotFound,
				errs.CodeUserNotFound, err,
			)
		}

		return nil, errs.New(
			errs.Internal,
			errs.CodeInternal, err,
		)
	}

	return openapi.GetUser200JSONResponse(toOpenAPIUser(usr)), nil
}

func (h *Handler) UpdateUser(
	ctx context.Context,
	request openapi.UpdateUserRequestObject,
) (openapi.UpdateUserResponseObject, error) {
	userUUID, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s", err,
		)
	}

	uu, err := parseUpdateUser(
		request.Body.Name,
		request.Body.Email,
		request.Body.Department,
		request.Body.Password,
		request.Body.PasswordConfirm,
		request.Body.Enabled,
	)
	if err != nil {
		return nil, errs.New(
			errs.InvalidArgument,
			errs.CodeValidation, err,
		)
	}

	currentUsr, err := h.svc.user.QueryByID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("update user: query: %w", err)
	}

	updUsr, err := h.svc.user.Update(ctx, currentUsr, uu)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return openapi.UpdateUser200JSONResponse(toOpenAPIUser(updUsr)), nil
}

func (h *Handler) DeleteUser(
	ctx context.Context,
	request openapi.DeleteUserRequestObject,
) (openapi.DeleteUserResponseObject, error) {
	userUUID, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s", err,
		)
	}

	currentUsr, err := h.svc.user.QueryByID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("delete user: query: %w", err)
	}

	err = h.svc.user.Delete(ctx, currentUsr)
	if err != nil {
		return nil, fmt.Errorf("delete user: %w", err)
	}

	return openapi.DeleteUser204Response{}, nil
}

func (h *Handler) CreateUserRole(
	ctx context.Context,
	request openapi.CreateUserRoleRequestObject,
) (openapi.CreateUserRoleResponseObject, error) {
	userUUID, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse user uuid: %s", err,
		)
	}

	roleUUID, err := uuid.Parse(request.Body.RoleId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse role uuid: %s", err,
		)
	}

	err = h.svc.userRoles.Add(ctx, userUUID, roleUUID)
	if err != nil {
		return nil, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"create user role: %s", err,
		)
	}

	return openapi.CreateUserRole204Response{}, nil
}

func (h *Handler) DeleteUserRole(
	ctx context.Context,
	request openapi.DeleteUserRoleRequestObject,
) (openapi.DeleteUserRoleResponseObject, error) {
	userUUID, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse user uuid: %s", err,
		)
	}

	roleUUID, err := uuid.Parse(request.Params.RoleId)
	if err != nil {
		return nil, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse role uuid: %s", err,
		)
	}

	err = h.svc.userRoles.Remove(ctx, userUUID, roleUUID)
	if err != nil {
		return nil, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"delete user role: %s", err,
		)
	}

	return openapi.DeleteUserRole204Response{}, nil
}

// ---------------------------------------------------------------------------
// Parsing helpers
// ---------------------------------------------------------------------------

func parseNewUser(body *openapi.NewUser) (user.NewUser, error) {
	var fieldErrors errs.FieldErrors

	addr, err := mail.ParseAddress(body.Email)
	if err != nil {
		fieldErrors.Add("email", err)
	}

	nme, err := name.Parse(body.Name)
	if err != nil {
		fieldErrors.Add("name", err)
	}

	dept := name.Null{}
	if body.Department != nil {
		dept, err = name.ParseNull(*body.Department)
		if err != nil {
			fieldErrors.Add("department", err)
		}
	}

	pass, err := password.ParseConfirm(body.Password, body.PasswordConfirm)
	if err != nil {
		fieldErrors.Add("password", err)
	}

	if len(fieldErrors) > 0 {
		return user.NewUser{}, fmt.Errorf("validate: %w", fieldErrors.ToError())
	}

	return user.NewUser{
		Name:       nme,
		Email:      *addr,
		Department: dept,
		Password:   pass,
	}, nil
}

func parseUpdateUser(
	namePtr, emailPtr, departmentPtr, passwordPtr, passwordConfirmPtr *string,
	enabledPtr *bool,
) (user.UpdateUser, error) {
	var fieldErrors errs.FieldErrors

	var uu user.UpdateUser

	if emailPtr != nil {
		addr, err := mail.ParseAddress(*emailPtr)
		if err != nil {
			fieldErrors.Add("email", err)
		} else {
			uu.Email = addr
		}
	}

	if namePtr != nil {
		nme, err := name.Parse(*namePtr)
		if err != nil {
			fieldErrors.Add("name", err)
		} else {
			uu.Name = &nme
		}
	}

	if departmentPtr != nil {
		dept, err := name.ParseNull(*departmentPtr)
		if err != nil {
			fieldErrors.Add("department", err)
		} else {
			uu.Department = &dept
		}
	}

	if passwordPtr != nil || passwordConfirmPtr != nil {
		pass, err := password.ParseConfirmPointers(passwordPtr, passwordConfirmPtr)
		if err != nil {
			fieldErrors.Add("password", err)
		} else {
			uu.Password = &pass
		}
	}

	uu.Enabled = enabledPtr

	if len(fieldErrors) > 0 {
		return user.UpdateUser{}, fmt.Errorf("validate: %w", fieldErrors.ToError())
	}

	return uu, nil
}

func parseUserFilter(id, nameStr, email, startCreatedDate, endCreatedDate string) (user.QueryFilter, error) {
	var (
		fieldErrors errs.FieldErrors
		filter      user.QueryFilter
	)

	if id != "" {
		uid, err := uuid.Parse(id)
		if err != nil {
			fieldErrors.Add("user_id", err)
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

	if email != "" {
		addr, err := mail.ParseAddress(email)
		if err != nil {
			fieldErrors.Add("email", err)
		} else {
			filter.Email = addr
		}
	}

	if startCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, startCreatedDate)
		if err != nil {
			fieldErrors.Add("start_created_date", err)
		} else {
			filter.StartCreatedDate = &t
		}
	}

	if endCreatedDate != "" {
		t, err := time.Parse(time.RFC3339, endCreatedDate)
		if err != nil {
			fieldErrors.Add("end_created_date", err)
		} else {
			filter.EndCreatedDate = &t
		}
	}

	if fieldErrors != nil {
		return user.QueryFilter{}, fieldErrors.ToError()
	}

	return filter, nil
}

func userDefaultOrderBy() order.By {
	return order.NewBy("user_id", order.ASC)
}

func userOrderByFields() map[string]string {
	return map[string]string{
		"user_id": user.OrderByID,
		"name":    user.OrderByName,
		"email":   user.OrderByEmail,
		"enabled": user.OrderByEnabled,
	}
}

func userIDExtractor(u user.User) string {
	return u.ID().String()
}

func userFieldExtractor(orderBy order.By) func(user.User) any {
	return func(u user.User) any {
		switch orderBy.Field {
		case user.OrderByName:
			return u.Name().String()
		case user.OrderByEmail:
			return u.Email().Address
		case user.OrderByEnabled:
			return u.Enabled()
		default:
			return u.ID().String()
		}
	}
}
