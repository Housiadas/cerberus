// Package user_usecase maintains the use case layer api the model
package user_usecase

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/Housiadas/cerberus/internal/core/domain/event"
	"github.com/Housiadas/cerberus/internal/core/domain/outbox"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/core/service/outbox_service"
	"github.com/Housiadas/cerberus/internal/core/service/user_service"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/internal/utils/page"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type UseCase struct {
	userCore  *user_service.Service
	outboxSvc *outbox_service.Service
	tx        pgsql.Beginner
}

func NewUseCase(
	userBus *user_service.Service,
	outboxSvc *outbox_service.Service,
	tx pgsql.Beginner,
) *UseCase {
	return &UseCase{
		userCore:  userBus,
		outboxSvc: outboxSvc,
		tx:        tx,
	}
}

// Create adds a new user to the system.
func (a *UseCase) Create(ctx context.Context, app NewUser) (User, error) {
	nc, err := toBusNewUser(app)
	if err != nil {
		return User{}, errs.New(errs.InvalidArgument, err)
	}

	var usr user.User
	txErr := pgsql.RunInTx(a.tx, func(tran pgsql.CommitRollbacker) error {
		userCoreTx, err := a.userCore.NewWithTx(tran)
		if err != nil {
			return errs.Errorf(errs.Internal, "user tx: %s", err)
		}

		outboxSvcTx, err := a.outboxSvc.NewWithTx(tran)
		if err != nil {
			return errs.Errorf(errs.Internal, "outbox tx: %s", err)
		}

		usr, err = userCoreTx.Create(ctx, nc)
		if err != nil {
			if errors.Is(err, user.ErrUniqueEmail) {
				return errs.New(errs.Aborted, user.ErrUniqueEmail)
			}

			return errs.Errorf(errs.Internal, "create: usr[%+v]: %s", usr, err)
		}

		return outboxSvcTx.Create(ctx, outbox.NewOutbox{
			EventType:   event.UserCreated,
			AggregateID: usr.ID(),
			Topic:       event.UserTopic,
			Payload:     usr,
		})
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
		return User{}, errs.New(errs.InvalidArgument, err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return User{}, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	currentUsr, err := a.userCore.QueryByID(ctx, userUUID)
	if err != nil {
		return User{}, errs.Errorf(
			errs.Internal,
			"query by id: userID[%s] uu[%+v]: %s",
			userUUID,
			uu,
			err,
		)
	}

	var updUsr user.User

	txErr := pgsql.RunInTx(a.tx, func(tran pgsql.CommitRollbacker) error {
		userCoreTx, initErr := a.userCore.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, "user tx: %s", initErr)
		}

		outboxSvcTx, initErr := a.outboxSvc.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, "outbox tx: %s", initErr)
		}

		var updateErr error

		updUsr, updateErr = userCoreTx.Update(ctx, currentUsr, uu)
		if updateErr != nil {
			return errs.Errorf(
				errs.Internal,
				"update: userID[%s] uu[%+v]: %s",
				userUUID,
				uu,
				updateErr,
			)
		}

		return outboxSvcTx.Create(ctx, outbox.NewOutbox{
			EventType:   event.UserUpdated,
			AggregateID: updUsr.ID(),
			Topic:       event.UserTopic,
			Payload:     updUsr,
		})
	})
	if txErr != nil {
		return User{}, fmt.Errorf("update user: %w", txErr)
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
		return errs.Errorf(
			errs.Internal,
			"query by id: userID[%s] uu[%+v]: %s",
			userUUID,
			currentUsr,
			err,
		)
	}

	err = a.userCore.Delete(ctx, currentUsr)
	if err != nil {
		return errs.Errorf(errs.Internal, "delete: userID[%s]: %s", userUUID, err)
	}

	err = a.outboxSvc.Create(ctx, outbox.NewOutbox{
		EventType:   event.UserDeleted,
		AggregateID: currentUsr.ID(),
		Topic:       event.UserTopic,
		Payload:     currentUsr,
	})
	if err != nil {
		return errs.Errorf(errs.Internal, "outbox create: %s", err)
	}

	return nil
}

// Query returns a list of users with paging.
func (a *UseCase) Query(ctx context.Context, qp AppQueryParams) (page.Result[User], error) {
	p, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return page.Result[User]{}, errs.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return page.Result[User]{}, err
	}

	orderBy, err := order.Parse(getOrderByFields(), qp.OrderBy, getDefaultOrderBy())
	if err != nil {
		return page.Result[User]{}, errs.NewFieldErrors("order", err)
	}

	usrs, err := a.userCore.Query(ctx, filter, orderBy, p)
	if err != nil {
		return page.Result[User]{}, errs.Errorf(errs.Internal, "query: %s", err)
	}

	total, err := a.userCore.Count(ctx, filter)
	if err != nil {
		return page.Result[User]{}, errs.Errorf(errs.Internal, "count: %s", err)
	}

	return page.NewResult(toAppUsers(usrs), total, p), nil
}

// QueryByID returns a user by its Ia.
func (a *UseCase) QueryByID(ctx context.Context, userID string) (User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return User{}, errs.Errorf(errs.InvalidArgument, "could not parse uuid: %s", err)
	}

	usr, err := a.userCore.QueryByID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return User{}, errs.New(errs.NotFound, err)
		}

		return User{}, errs.New(errs.Internal, err)
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
			return User{}, errs.New(errs.NotFound, err)
		}

		return User{}, fmt.Errorf("user authenticate error: %w", err)
	}

	return toAppUser(usr), nil
}
