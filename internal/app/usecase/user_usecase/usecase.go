// Package user_usecase maintains the use case layer api the model
package user_usecase

import (
	"context"
	"errors"
	"net/mail"

	"github.com/Housiadas/cerberus/internal/common/validation"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/web"
	"github.com/Housiadas/cerberus/pkg/web/errs"
	"github.com/google/uuid"
)

// UseCase manages the set of cli layer api functions for the user core.
type UseCase struct {
	userCore *user_service.Service
}

// NewUseCase constructs a user cli API for use.
func NewUseCase(userBus *user_service.Service) *UseCase {
	return &UseCase{
		userCore: userBus,
	}
}

// Create adds a new user to the system.
func (a *UseCase) Create(ctx context.Context, app NewUser) (User, error) {
	nc, err := toBusNewUser(app)
	if err != nil {
		return User{}, errs.New(errs.InvalidArgument, err)
	}

	usr, err := a.userCore.Create(ctx, nc)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmail) {
			return User{}, errs.New(errs.Aborted, user.ErrUniqueEmail)
		}
		return User{}, errs.Errorf(errs.Internal, "create: usr[%+v]: %s", usr, err)
	}

	return toAppUser(usr), nil
}

// Update updates an existing user.
func (a *UseCase) Update(ctx context.Context, res UpdateUser, userID string) (User, error) {
	uu, err := toBusUpdateUser(res)
	if err != nil {
		return User{}, errs.New(errs.InvalidArgument, err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return User{}, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	currentUsr, err := a.userCore.QueryByID(ctx, userUUID)
	if err != nil {
		return User{}, errs.Errorf(errs.Internal, "query by id: userID[%s] uu[%+v]: %s", userUUID, uu, err)
	}

	updUsr, err := a.userCore.Update(ctx, currentUsr, uu)
	if err != nil {
		return User{}, errs.Errorf(errs.Internal, "update: userID[%s] uu[%+v]: %s", userUUID, uu, err)
	}

	return toAppUser(updUsr), nil
}

// Delete removes a user from the system.
func (a *UseCase) Delete(ctx context.Context, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	currentUsr, err := a.userCore.QueryByID(ctx, userUUID)
	if err != nil {
		return errs.Errorf(errs.Internal, "query by id: userID[%s] uu[%+v]: %s", userUUID, currentUsr, err)
	}

	if err := a.userCore.Delete(ctx, currentUsr); err != nil {
		return errs.Errorf(errs.Internal, "delete: userID[%s]: %s", userUUID, err)
	}

	return nil
}

// Query returns a list of users with paging.
func (a *UseCase) Query(ctx context.Context, qp AppQueryParams) (web.Result[User], error) {
	p, err := web.Parse(qp.Page, qp.Rows)
	if err != nil {
		return web.Result[User]{}, validation.ErrorfieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return web.Result[User]{}, err
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return web.Result[User]{}, validation.ErrorfieldErrors("order", err)
	}

	usrs, err := a.userCore.Query(ctx, filter, orderBy, p)
	if err != nil {
		return web.Result[User]{}, errs.Errorf(errs.Internal, "query: %s", err)
	}

	total, err := a.userCore.Count(ctx, filter)
	if err != nil {
		return web.Result[User]{}, errs.Errorf(errs.Internal, "count: %s", err)
	}

	return web.NewResult(toAppUsers(usrs), total, p), nil
}

// QueryByID returns a user by its Ia.
func (a *UseCase) QueryByID(ctx context.Context, userID string) (User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return User{}, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	usr, err := a.userCore.QueryByID(ctx, userUUID)
	if err != nil {
		return User{}, errs.Errorf(errs.Internal, "query_by_id: %s", err)
	}

	return toAppUser(usr), nil
}

// Authenticate provides an API to authenticate the user.
func (a *UseCase) Authenticate(ctx context.Context, authUser AuthenticateUser) (User, error) {
	addr, err := mail.ParseAddress(authUser.Email)
	if err != nil {
		return User{}, validation.ErrorfieldErrors("email", err)
	}

	usr, err := a.userCore.Authenticate(ctx, *addr, authUser.Password)
	if err != nil {
		return User{}, err
	}

	return toAppUser(usr), nil
}
