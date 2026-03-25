// Package user_usecase maintains the use case layer api the model
package user_usecase

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/core/user/user_service"
	"github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/eventbus"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type UseCase struct {
	log        logger.Logger
	tx         pgsql.Beginner
	userCore   *user_service.Service
	dispatcher *eventbus.EventDispatcher
}

func NewUseCase(
	log logger.Logger,
	userBus *user_service.Service,
	dispatcher *eventbus.EventDispatcher,
	tx pgsql.Beginner,
) *UseCase {
	return &UseCase{
		log:        log,
		tx:         tx,
		userCore:   userBus,
		dispatcher: dispatcher,
	}
}

// Create adds a new user to the system.
func (a *UseCase) Create(ctx context.Context, app NewUser) (User, error) {
	nc, err := toBusNewUser(app)
	if err != nil {
		return User{}, errs.New(errs.InvalidArgument, errs.CodeValidation, err)
	}

	var usr user.User

	txErr := pgsql.RunInTx(ctx, a.log, a.tx, func(tran pgsql.CommitRollbacker) error {
		userCoreTx, err := a.userCore.NewWithTx(tran)
		if err != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "user tx: %s", err)
		}

		usr, err = userCoreTx.Create(ctx, nc)
		if err != nil {
			if errors.Is(err, user.ErrUniqueEmail) {
				return errs.New(errs.Aborted, errs.CodeUniqueEmail, user.ErrUniqueEmail)
			}

			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
				"create: usr[%+v]: %s",
				usr,
				err,
			)
		}

		return a.dispatcher.Dispatch(
			ctx,
			tran,
			newUserEvent(usr, event.UserCreated, audit.ActionCreate),
		)
	})
	if txErr != nil {
		return User{}, fmt.Errorf("create user: %w", txErr)
	}

	return toAppUser(usr), nil
}

// Update updates an existing user.
func (a *UseCase) Update(ctx context.Context, res UpdateUser, userID string) (User, error) {
	uu, err := toBusUpdateUser(res)
	if err != nil {
		return User{}, errs.New(errs.InvalidArgument, errs.CodeValidation, err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return User{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentUsr, err := a.userCore.QueryByID(ctx, userUUID)
	if err != nil {
		return User{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"query by id: userID[%s] uu[%+v]: %s",
			userUUID,
			uu,
			err,
		)
	}

	var updUsr user.User

	txErr := pgsql.RunInTx(ctx, a.log, a.tx, func(tran pgsql.CommitRollbacker) error {
		var runErr error

		updUsr, runErr = a.updateInTx(ctx, tran, currentUsr, uu, userUUID)

		return runErr
	})
	if txErr != nil {
		return User{}, fmt.Errorf("update user: %w", txErr)
	}

	return toAppUser(updUsr), nil
}

// UpdateMe updates the authenticated user's own profile.
func (a *UseCase) UpdateMe(ctx context.Context, res UpdateMe, userID string) (User, error) {
	return a.Update(ctx, UpdateUser{
		Name:            res.Name,
		Email:           res.Email,
		Department:      res.Department,
		Password:        res.Password,
		PasswordConfirm: res.PasswordConfirm,
	}, userID)
}

// Delete removes a user from the system.
func (a *UseCase) Delete(ctx context.Context, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentUsr, err := a.userCore.QueryByID(ctx, userUUID)
	if err != nil {
		return errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"query by id: userID[%s] uu[%+v]: %s",
			userUUID,
			currentUsr,
			err,
		)
	}

	txErr := pgsql.RunInTx(ctx, a.log, a.tx, func(tran pgsql.CommitRollbacker) error {
		deleteErr := a.userCore.Delete(ctx, currentUsr)
		if deleteErr != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
				"delete: userID[%s]: %s",
				userUUID,
				deleteErr,
			)
		}

		return a.dispatcher.Dispatch(
			ctx,
			tran,
			newUserEvent(currentUsr, event.UserDeleted, audit.ActionDelete),
		)
	})
	if txErr != nil {
		return fmt.Errorf("delete user: %w", txErr)
	}

	return nil
}

// Query returns a list of users with cursor-based paging.
func (a *UseCase) Query(ctx context.Context, qp AppQueryParams) (cursor.Result[User], error) {
	cur, err := cursor.Parse(qp.Cursor, qp.Limit)
	if err != nil {
		return cursor.Result[User]{}, errs.NewFieldErrors("cursor", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return cursor.Result[User]{}, err
	}

	orderBy, err := order.Parse(getOrderByFields(), qp.OrderBy, getDefaultOrderBy())
	if err != nil {
		return cursor.Result[User]{}, errs.NewFieldErrors("order", err)
	}

	usrs, err := a.userCore.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return cursor.Result[User]{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"query: %s",
			err,
		)
	}

	return cursor.NewResult(
		toAppUsers(usrs),
		cur.Limit(),
		cur,
		orderBy,
		func(u User) string { return u.ID },
		userFieldExtractor(orderBy),
	), nil
}

// QueryByEmail returns a user by its email address.
func (a *UseCase) QueryByEmail(ctx context.Context, email string) (User, error) {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return User{}, errs.NewFieldErrors("email", err)
	}

	usr, err := a.userCore.QueryByEmail(ctx, *addr)
	if err != nil {
		return User{}, errs.New(errs.NotFound, errs.CodeUserNotFound, err)
	}

	return toAppUser(usr), nil
}

// QueryByID returns a user by its Ia.
func (a *UseCase) QueryByID(ctx context.Context, userID string) (User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return User{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	usr, err := a.userCore.QueryByID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return User{}, errs.New(errs.NotFound, errs.CodeUserNotFound, err)
		}

		return User{}, errs.New(errs.Internal, errs.CodeInternal, err)
	}

	return toAppUser(usr), nil
}

// Authenticate provides an API to authenticate the user.
func (a *UseCase) Authenticate(ctx context.Context, authUser AuthenticateUser) (User, error) {
	addr, err := mail.ParseAddress(authUser.Email)
	if err != nil {
		return User{}, errs.NewFieldErrors("email", err)
	}

	usr, err := a.userCore.Authenticate(ctx, *addr, authUser.Password)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return User{}, errs.New(errs.NotFound, errs.CodeAuthFailed, err)
		}

		return User{}, fmt.Errorf("user authenticate error: %w", err)
	}

	return toAppUser(usr), nil
}

func (a *UseCase) updateInTx(
	ctx context.Context,
	tran pgsql.CommitRollbacker,
	currentUsr user.User,
	uu user.UpdateUser,
	userUUID uuid.UUID,
) (user.User, error) {
	userCoreTx, initErr := a.userCore.NewWithTx(tran)
	if initErr != nil {
		return user.User{}, errs.Errorf(errs.Internal, errs.CodeInternal, "user tx: %s", initErr)
	}

	updUsr, updateErr := userCoreTx.Update(ctx, currentUsr, uu)
	if updateErr != nil {
		return user.User{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"update: userID[%s] uu[%+v]: %s",
			userUUID,
			uu,
			updateErr,
		)
	}

	dispatchErr := a.dispatcher.Dispatch(
		ctx,
		tran,
		newUserEvent(updUsr, event.UserUpdated, audit.ActionUpdate),
	)
	if dispatchErr != nil {
		return user.User{}, fmt.Errorf("dispatch: %w", dispatchErr)
	}

	return updUsr, nil
}

func newUserEvent(usr user.User, eventType string, action string) event.DomainEvent {
	return event.DomainEvent{
		EventType:   eventType,
		AggregateID: usr.ID(),
		Topic:       event.UserTopic,
		Payload:     usr,
		ObjEntity:   entity.New(entity.UserEntity),
		ObjName:     usr.Name(),
		Action:      action,
		Message:     "user " + action,
	}
}
