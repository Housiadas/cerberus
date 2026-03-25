// Package permission_usecase maintains the usecase layer api the model
package permission_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/audit"
	"github.com/Housiadas/cerberus/internal/core/permission"
	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/eventbus"
	"github.com/Housiadas/cerberus/internal/types/entity"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Debugc(ctx context.Context, caller int, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Infoc(ctx context.Context, caller int, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Warnc(ctx context.Context, caller int, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Errorc(ctx context.Context, caller int, msg string, args ...any)
}

type UseCase struct {
	log               logger
	permissionService *permission.Service
	dispatcher        *eventbus.EventDispatcher
	tx                pgsql.Beginner
}

func NewUseCase(
	log logger,
	permissionService *permission.Service,
	dispatcher *eventbus.EventDispatcher,
	tx pgsql.Beginner,
) *UseCase {
	return &UseCase{
		log:               log,
		tx:                tx,
		dispatcher:        dispatcher,
		permissionService: permissionService,
	}
}

// Create adds a new permission to the system.
func (uc *UseCase) Create(ctx context.Context, nperm NewPermission) (Permission, error) {
	np, err := toBusNewPermission(nperm)
	if err != nil {
		return Permission{}, errs2.New(errs2.InvalidArgument, errs2.CodeValidation, err)
	}

	var perm permission.Permission

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		permServiceTx, initErr := uc.permissionService.NewWithTx(tran)
		if initErr != nil {
			return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "permission tx: %s", initErr)
		}

		perm, err = permServiceTx.Create(ctx, np)
		if err != nil {
			return errs2.Errorf(
				errs2.Internal,
				errs2.CodeInternal,
				"create: perm[%+v]: %s",
				perm,
				err,
			)
		}

		return uc.dispatcher.Dispatch(ctx, tran, newPermissionEvent(perm, audit.ActionCreate))
	})
	if txErr != nil {
		return Permission{}, fmt.Errorf("create permission: %w", txErr)
	}

	return toAppPermission(perm), nil
}

// Update updates an existing permission.
func (uc *UseCase) Update(
	ctx context.Context,
	res UpdatePermission,
	permissionID string,
) (Permission, error) {
	up, err := toBusUpdatePermission(res)
	if err != nil {
		return Permission{}, errs2.New(errs2.InvalidArgument, errs2.CodeValidation, err)
	}

	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return Permission{}, errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentPerm, err := uc.permissionService.QueryByID(ctx, permissionUUID)
	if err != nil {
		return Permission{}, errs2.Errorf(
			errs2.Internal,
			errs2.CodeInternal,
			"permission query by id: %s",
			err,
		)
	}

	var updPerm permission.Permission

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		permServiceTx, initErr := uc.permissionService.NewWithTx(tran)
		if initErr != nil {
			return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "permission tx: %s", initErr)
		}

		updPerm, err = permServiceTx.Update(ctx, currentPerm, up)
		if err != nil {
			return errs2.Errorf(
				errs2.Internal,
				errs2.CodeInternal,
				"update: permissionID[%s] up[%+v]: %s",
				permissionID,
				up,
				err,
			)
		}

		return uc.dispatcher.Dispatch(ctx, tran, newPermissionEvent(updPerm, audit.ActionUpdate))
	})
	if txErr != nil {
		return Permission{}, fmt.Errorf("update permission: %w", txErr)
	}

	return toAppPermission(updPerm), nil
}

// Delete removes a permission from the system.
func (uc *UseCase) Delete(ctx context.Context, permissionID string) error {
	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentPerm, err := uc.permissionService.QueryByID(ctx, permissionUUID)
	if err != nil {
		return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "permission query by id: %s", err)
	}

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		permServiceTx, initErr := uc.permissionService.NewWithTx(tran)
		if initErr != nil {
			return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "permission tx: %s", initErr)
		}

		deleteErr := permServiceTx.Delete(ctx, currentPerm)
		if deleteErr != nil {
			return errs2.Errorf(
				errs2.Internal,
				errs2.CodeInternal,
				"delete: permissionID[%s]: %s",
				permissionUUID,
				deleteErr,
			)
		}

		return uc.dispatcher.Dispatch(
			ctx,
			tran,
			newPermissionEvent(currentPerm, audit.ActionDelete),
		)
	})
	if txErr != nil {
		return fmt.Errorf("delete permission: %w", txErr)
	}

	return nil
}

// QueryByID returns a permission by its ID.
func (uc *UseCase) QueryByID(ctx context.Context, permissionID string) (Permission, error) {
	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return Permission{}, errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	perm, err := uc.permissionService.QueryByID(ctx, permissionUUID)
	if err != nil {
		return Permission{}, errs2.Errorf(
			errs2.Internal,
			errs2.CodeInternal,
			"permission query by id: %s",
			err,
		)
	}

	return toAppPermission(perm), nil
}

// Query returns a list of permissions with cursor-based paging.
func (uc *UseCase) Query(
	ctx context.Context,
	qp AppQueryParams,
) (cursor.Result[Permission], error) {
	cur, err := cursor.Parse(qp.Cursor, qp.Limit)
	if err != nil {
		return cursor.Result[Permission]{}, errs2.NewFieldErrors("cursor", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return cursor.Result[Permission]{}, err
	}

	orderBy, err := order.Parse(getOrderByFields(), qp.OrderBy, getDefaultOrderBy())
	if err != nil {
		return cursor.Result[Permission]{}, errs2.NewFieldErrors("order", err)
	}

	perms, err := uc.permissionService.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return cursor.Result[Permission]{}, errs2.Errorf(
			errs2.Internal,
			errs2.CodeInternal,
			"query: %s",
			err,
		)
	}

	return cursor.NewResult(
		toAppPermissions(perms),
		cur.Limit(),
		cur,
		orderBy,
		func(p Permission) string { return p.ID },
		permissionFieldExtractor(orderBy),
	), nil
}

func newPermissionEvent(p permission.Permission, action string) event.DomainEvent {
	return event.DomainEvent{
		AggregateID: p.ID(),
		Payload:     p,
		ObjEntity:   entity.New(entity.PermissionEntity),
		ObjName:     p.Name(),
		Action:      action,
		Message:     "permission " + action,
	}
}
