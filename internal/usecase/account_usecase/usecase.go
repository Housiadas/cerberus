// Package account_usecase maintains the use case layer for accounts
package account_usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/account"
	"github.com/Housiadas/cerberus/internal/core/service/account_service"
	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/pkg/cursor"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/order"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/google/uuid"
)

// UseCase manages the set of APIs for account use cases.
type UseCase struct {
	log        logger.Logger
	tx         pgsql.Beginner
	accountSvc *account_service.Service
}

// NewUseCase constructs a use case for account operations.
func NewUseCase(
	log logger.Logger,
	accountSvc *account_service.Service,
	tx pgsql.Beginner,
) *UseCase {
	return &UseCase{
		log:        log,
		tx:         tx,
		accountSvc: accountSvc,
	}
}

// Create adds a new account to the system.
func (uc *UseCase) Create(ctx context.Context, app NewAccount) (Account, error) {
	na := account.NewAccount{
		Name: app.Name,
	}

	var acc account.Account

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		svcTx, err := uc.accountSvc.NewWithTx(tran)
		if err != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "account tx: %s", err)
		}

		acc, err = svcTx.Create(ctx, na)
		if err != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "create: %s", err)
		}

		return nil
	})
	if txErr != nil {
		return Account{}, fmt.Errorf("create account: %w", txErr)
	}

	return toAppAccount(acc), nil
}

// Update updates an existing account.
func (uc *UseCase) Update(
	ctx context.Context,
	app UpdateAccount,
	accountID string,
) (Account, error) {
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return Account{}, errs.Errorf(
			errs.InvalidArgument, errs.CodeValidation, "could not parse uuid: %s", err,
		)
	}

	currentAcc, err := uc.accountSvc.QueryByID(ctx, accountUUID)
	if err != nil {
		return Account{}, errs.Errorf(
			errs.Internal, errs.CodeInternal, "query by id: accountID[%s]: %s", accountUUID, err,
		)
	}

	ua := account.UpdateAccount{
		Name: app.Name,
	}

	var updAcc account.Account

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		svcTx, initErr := uc.accountSvc.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "account tx: %s", initErr)
		}

		var updateErr error

		updAcc, updateErr = svcTx.Update(ctx, currentAcc, ua)
		if updateErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "update: %s", updateErr)
		}

		return nil
	})
	if txErr != nil {
		return Account{}, fmt.Errorf("update account: %w", txErr)
	}

	return toAppAccount(updAcc), nil
}

// Delete removes an account from the system.
func (uc *UseCase) Delete(ctx context.Context, accountID string) error {
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return errs.Errorf(
			errs.InvalidArgument, errs.CodeValidation, "could not parse uuid: %s", err,
		)
	}

	currentAcc, err := uc.accountSvc.QueryByID(ctx, accountUUID)
	if err != nil {
		return errs.Errorf(
			errs.Internal, errs.CodeInternal, "query by id: accountID[%s]: %s", accountUUID, err,
		)
	}

	txErr := pgsql.RunInTx(ctx, uc.log, uc.tx, func(tran pgsql.CommitRollbacker) error {
		svcTx, initErr := uc.accountSvc.NewWithTx(tran)
		if initErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "account tx: %s", initErr)
		}

		deleteErr := svcTx.Delete(ctx, currentAcc)
		if deleteErr != nil {
			return errs.Errorf(errs.Internal, errs.CodeInternal, "delete: %s", deleteErr)
		}

		return nil
	})
	if txErr != nil {
		return fmt.Errorf("delete account: %w", txErr)
	}

	return nil
}

// Query returns a list of accounts with cursor-based paging.
func (uc *UseCase) Query(ctx context.Context, qp AppQueryParams) (cursor.Result[Account], error) {
	cur, err := cursor.Parse(qp.Cursor, qp.Limit)
	if err != nil {
		return cursor.Result[Account]{}, errs.NewFieldErrors("cursor", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return cursor.Result[Account]{}, err
	}

	orderBy, err := order.Parse(getOrderByFields(), qp.OrderBy, getDefaultOrderBy())
	if err != nil {
		return cursor.Result[Account]{}, errs.NewFieldErrors("order", err)
	}

	accs, err := uc.accountSvc.Query(ctx, filter, orderBy, cur)
	if err != nil {
		return cursor.Result[Account]{}, errs.Errorf(
			errs.Internal,
			errs.CodeInternal,
			"query: %s",
			err,
		)
	}

	return cursor.NewResult(
		toAppAccounts(accs),
		cur.Limit(),
		cur,
		orderBy,
		func(a Account) string { return a.ID },
		accountFieldExtractor(orderBy),
	), nil
}

// QueryByID returns an account by its ID.
func (uc *UseCase) QueryByID(ctx context.Context, accountID string) (Account, error) {
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return Account{}, errs.Errorf(
			errs.InvalidArgument, errs.CodeValidation, "could not parse uuid: %s", err,
		)
	}

	acc, err := uc.accountSvc.QueryByID(ctx, accountUUID)
	if err != nil {
		if errors.Is(err, account.ErrNotFound) {
			return Account{}, errs.New(errs.NotFound, errs.CodeAccountNotFound, err)
		}

		return Account{}, errs.New(errs.Internal, errs.CodeInternal, err)
	}

	return toAppAccount(acc), nil
}
