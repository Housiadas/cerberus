// Package permission_usecase maintains the usecase layer api the model
package permission_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/app/event_dispatcher"
	"github.com/Housiadas/cerberus/internal/core/domain/audit"
	"github.com/Housiadas/cerberus/internal/core/domain/entity"
	"github.com/Housiadas/cerberus/internal/core/domain/event"
	"github.com/Housiadas/cerberus/internal/core/domain/permission"
	"github.com/Housiadas/cerberus/internal/core/service/permission_service"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

type UseCase struct {
	log               logger.Logger
	permissionService *permission_service.Service
	dispatcher        *event_dispatcher.EventDispatcher
	tx                pgsql.Beginner
}

func NewUseCase(
	log logger.Logger,
	permissionService *permission_service.Service,
	dispatcher *event_dispatcher.EventDispatcher,
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
		return Permission{}, errs.New(errs.InvalidArgument, errs.CodeValidation, err)
	}

	var perm permission.Permission

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		permServiceTx, initErr := uc.permissionService.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "permission tx: %s", initErr)
		}

		perm, err = permServiceTx.Create(ctx, np)
		if err != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
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
		return Permission{}, errs.New(errs.InvalidArgument, errs.CodeValidation, err)
	}

	permissionUUID, err := uuid.Parse(permissionID)
	if err != nil {
		return Permission{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentPerm, err := uc.permissionService.QueryByID(ctx, permissionUUID)
	if err != nil {
		return Permission{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"permission query by id: %s",
			err,
		)
	}

	var updPerm permission.Permission

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		permServiceTx, initErr := uc.permissionService.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "permission tx: %s", initErr)
		}

		updPerm, err = permServiceTx.Update(ctx, currentPerm, up)
		if err != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
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
		return errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	currentPerm, err := uc.permissionService.QueryByID(ctx, permissionUUID)
	if err != nil {
		return errs.Errorf(errs.Internal, errs.CodeInternal, "permission query by id: %s", err)
	}

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		permServiceTx, initErr := uc.permissionService.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "permission tx: %s", initErr)
		}

		deleteErr := permServiceTx.Delete(ctx, currentPerm)
		if deleteErr != nil {
			return errs.Errorf(
				errs.Internal,
				errs.CodeInternal,
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
		return Permission{}, errs.Errorf(
			errs.InvalidArgument,
			errs.CodeValidation,
			"could not parse uuid: %s",
			err,
		)
	}

	perm, err := uc.permissionService.QueryByID(ctx, permissionUUID)
	if err != nil {
		return Permission{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"permission query by id: %s",
			err,
		)
	}

	return toAppPermission(perm), nil
}

// Query returns a list of permissions with cursor-based paging.
func (uc *UseCase) Query(ctx context.Context, qp AppQueryParams) (cursor.Result[Permission], error) {
	cur, err := cursor.Parse(qp.Cursor, qp.Limit)
	if err != nil {
		return cursor.Result[Permission]{}, errs.NewFieldErrors("cursor", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return cursor.Result[Permission]{}, err
	}

	orderBy, err := order.Parse(getOrderByFields(), qp.OrderBy, getDefaultOrderBy())
	if err != nil {
		return cursor.Result[Permission]{}, errs.NewFieldErrors("order", err)
	}

	perms, err := uc.permissionService.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return cursor.Result[Permission]{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
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
		func(p Permission) any { return p.ID },
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
